package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// WorkPoolSourceModel defines the Terraform data source model.
type WorkPoolSourceModel struct {
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

//nolint:ireturn
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

	d.client = client.WorkPools()
}

// Schema defines the schema for the data source.
func (d *WorkPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Work pool UUID",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of the work pool creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time that the work pool was last updated un RFC 3339 format",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the work pool",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the work pool",
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
			},
			"default_queue_id": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the default queue associated with this work pool",
			},
			"base_job_template": schema.StringAttribute{
				Computed:    true,
				Description: "The base job template for the work pool, as a JSON string",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkPoolSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := d.client.Get(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("error", err.Error())
	}

	data.ID = types.StringValue(pool.ID.String())

	if pool.Created == nil {
		data.Created = types.StringNull()
	} else {
		data.Created = types.StringValue(pool.Created.Format(time.RFC3339))
	}

	if pool.Updated == nil {
		data.Updated = types.StringNull()
	} else {
		data.Updated = types.StringValue(pool.Updated.Format(time.RFC3339))
	}

	data.Description = types.StringValue(pool.Name)
	data.Type = types.StringValue(pool.Type)
	data.Paused = types.BoolValue(pool.IsPaused)
	data.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	data.DefaultQueueID = types.StringValue(pool.DefaultQueueID.String())

	if pool.BaseJobTemplate != nil {
		var sb strings.Builder
		if err := json.NewEncoder(&sb).Encode(pool.BaseJobTemplate); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to serialize Base Job Template",
				fmt.Sprintf("Failed to serialize Base Job Template as JSON string: %s", err),
			)

			return
		}

		data.BaseJobTemplate = types.StringValue(sb.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
