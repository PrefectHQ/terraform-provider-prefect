package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = datasource.DataSourceWithConfigure(&WorkPoolsDataSource{})

// WorkPoolsDataSource contains state for the data source.
type WorkPoolsDataSource struct {
	client api.WorkPoolsClient
}

// WorkPoolsSourceModel defines the Terraform data source model.
type WorkPoolsSourceModel struct {
	FilterAny types.List `tfsdk:"filter_any"`
	WorkPools types.List `tfsdk:"work_pools"`
}

// NewWorkPoolsDataSource returns a new WorkPoolsDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkPoolsDataSource() datasource.DataSource {
	return &WorkPoolsDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkPoolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pools"
}

// Configure initializes runtime state for the data source.
func (d *WorkPoolsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema defines the schema for the data source.
func (d *WorkPoolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source for querying work pools",
		Attributes: map[string]schema.Attribute{
			"filter_any": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Work pool UUIDs to search for (work pools with any matching UUID are returned)",
			},
			"work_pools": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Work pools returned by the server",
				NestedObject: schema.NestedAttributeObject{
					Attributes: workPoolAttributes,
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkPoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkPoolsSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := api.WorkPoolFilter{}

	pools, err := d.client.List(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work pool state",
			fmt.Sprintf("Could not read work pool, unexpected error: %s", err.Error()),
		)

		return
	}

	attributeTypes := map[string]attr.Type{
		"id":                types.StringType,
		"created":           types.StringType,
		"updated":           types.StringType,
		"name":              types.StringType,
		"description":       types.StringType,
		"type":              types.StringType,
		"paused":            types.BoolType,
		"concurrency_limit": types.Int64Type,
		"default_queue_id":  types.StringType,
		"base_job_template": types.StringType,
	}

	poolObjects := make([]attr.Value, 0, len(pools))
	for _, pool := range pools {
		attributeValues := map[string]attr.Value{
			"id":                types.StringValue(pool.ID.String()),
			"name":              types.StringValue(pool.Name),
			"description":       types.StringPointerValue(pool.Description),
			"type":              types.StringValue(pool.Type),
			"paused":            types.BoolValue(pool.IsPaused),
			"concurrency_limit": types.Int64PointerValue(pool.ConcurrencyLimit),
			"default_queue_id":  types.StringValue(pool.DefaultQueueID.String()),
		}

		if pool.Created == nil {
			attributeValues["created"] = types.StringNull()
		} else {
			attributeValues["created"] = types.StringValue(pool.Created.Format(time.RFC3339))
		}

		if pool.Updated == nil {
			attributeValues["updated"] = types.StringNull()
		} else {
			attributeValues["updated"] = types.StringValue(pool.Updated.Format(time.RFC3339))
		}

		if pool.BaseJobTemplate == nil {
			attributeValues["base_job_template"] = types.StringNull()
		} else {
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

			attributeValues["base_job_template"] = types.StringValue(builder.String())
		}

		poolObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		poolObjects = append(poolObjects, poolObject)
	}

	list, diag := types.ListValue(types.ObjectType{AttrTypes: attributeTypes}, poolObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.WorkPools = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
