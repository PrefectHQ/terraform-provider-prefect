package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = datasource.DataSourceWithConfigure(&WorkPoolDataSource{})

// WorkPoolDataSource contains state for the data source.
type WorkPoolDataSource struct {
	client api.WorkPoolsClient
}

// WorkPoolDataSourceModel defines the Terraform data source model.
type WorkPoolDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`

	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Type             types.String `tfsdk:"type"`
	Paused           types.Bool   `tfsdk:"paused"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
	DefaultQueueID   types.String `tfsdk:"default_queue_id"`
	BaseJobTemplate  types.String `tfsdk:"base_job_template"`
}

// NewWorkPoolDataSource returns a new WorkPoolDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkPoolDataSource() datasource.DataSource {
	return &WorkPoolDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pool"
}

// Configure initializes runtime state for the data source.
func (d *WorkPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client, _ = client.WorkPools(uuid.Nil, uuid.Nil)
}

var workPoolAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		Description: "Work pool UUID",
		Optional:    true,
	},
	"created": schema.StringAttribute{
		Computed:    true,
		Description: "Date and time of the work pool creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		Description: "Date and time that the work pool was last updated in RFC 3339 format",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name of the work pool",
		Optional:    true,
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of the work pool",
		Optional:    true,
	},
	"type": schema.StringAttribute{
		Computed:    true,
		Description: "Type of the work pool",
	},
	"paused": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether this work pool is paused",
	},
	"concurrency_limit": schema.Int64Attribute{
		Computed:    true,
		Description: "The concurrency limit applied to this work pool",
		Optional:    true,
	},
	"default_queue_id": schema.StringAttribute{
		Computed:    true,
		Description: "The UUID of the default queue associated with this work pool",
		Optional:    true,
	},
	"base_job_template": schema.StringAttribute{
		Computed:    true,
		Description: "The base job template for the work pool, as a JSON string",
	},
}

// Schema defines the schema for the data source.
func (d *WorkPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect Work Pool",
		Attributes:  workPoolAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkPoolDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := d.client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work pool state",
			fmt.Sprintf("Could not read work pool, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = types.StringValue(pool.ID.String())

	if pool.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(pool.Created.Format(time.RFC3339))
	}

	if pool.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(pool.Updated.Format(time.RFC3339))
	}

	model.Description = types.StringPointerValue(pool.Description)
	model.Type = types.StringValue(pool.Type)
	model.Paused = types.BoolValue(pool.IsPaused)
	model.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	model.DefaultQueueID = types.StringValue(pool.DefaultQueueID.String())

	if pool.BaseJobTemplate != nil {
		var builder strings.Builder
		encoder := json.NewEncoder(&builder)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(pool.BaseJobTemplate); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to serialize Base Job Template",
				fmt.Sprintf("Failed to serialize Base Job Template as JSON string: %s", err),
			)

			return
		}

		model.BaseJobTemplate = types.StringValue(builder.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
