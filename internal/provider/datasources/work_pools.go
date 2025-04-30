package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = datasource.DataSourceWithConfigure(&WorkPoolsDataSource{})

// WorkPoolsDataSource contains state for the data source.
type WorkPoolsDataSource struct {
	client api.PrefectClient
}

// WorkPoolsSourceModel defines the Terraform data source model.
type WorkPoolsSourceModel struct {
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}

// Schema defines the schema for the data source.
func (d *WorkPoolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an multiple Work Pools.
<br>
Use this data source to search for multiple Work Pools. Defaults to fetching all Work Pools in the Workspace.
<br>
For more information, see [configure dynamic infrastructure with work pools](https://docs.prefect.io/v3/deploy/infrastructure-concepts/work-pools).
`,
			helpers.AllPlans...,
		),
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
				Optional:    true,
			},
			"filter_any": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Work pool IDs (UUID) to search for (work pools with any matching UUID are returned)",
			},
			"work_pools": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Work pools returned by the server",
				NestedObject: schema.NestedAttributeObject{
					Attributes: workPoolAttributesBase,
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

	client, err := d.client.WorkPools(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool", err))

		return
	}

	filters := make([]string, 0, len(model.FilterAny.Elements()))

	for _, filter := range model.FilterAny.Elements() {
		u, err := uuid.Parse(filter.String())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("filter_any"),
				"Error parsing work pool %s ID",
				fmt.Sprintf("Could not parse work pool ID %s to UUID, unexpected error: %s", u.String(), err.Error()),
			)

			return
		}

		filters = append(filters, u.String())
	}

	pools, err := client.List(ctx, filters)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pools", "list", err))

		return
	}

	attributeTypes := map[string]attr.Type{
		"id":                customtypes.UUIDType{},
		"created":           customtypes.TimestampType{},
		"updated":           customtypes.TimestampType{},
		"name":              types.StringType,
		"description":       types.StringType,
		"type":              types.StringType,
		"paused":            types.BoolType,
		"concurrency_limit": types.Int64Type,
		"default_queue_id":  customtypes.UUIDType{},
		"base_job_template": types.StringType,
	}

	poolObjects := make([]attr.Value, 0, len(pools))
	for _, pool := range pools {
		attributeValues := map[string]attr.Value{
			"id":                customtypes.NewUUIDValue(pool.ID),
			"created":           customtypes.NewTimestampPointerValue(pool.Created),
			"updated":           customtypes.NewTimestampPointerValue(pool.Updated),
			"name":              types.StringValue(pool.Name),
			"description":       types.StringPointerValue(pool.Description),
			"type":              types.StringValue(pool.Type),
			"paused":            types.BoolValue(pool.IsPaused),
			"concurrency_limit": types.Int64PointerValue(pool.ConcurrencyLimit),
			"default_queue_id":  customtypes.NewUUIDValue(pool.DefaultQueueID),
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
