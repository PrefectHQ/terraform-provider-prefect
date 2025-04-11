package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

var _ = datasource.DataSourceWithConfigure(&globalConcurrencyLimitDataSource{})

type globalConcurrencyLimitDataSource struct {
	client api.PrefectClient
}

type globalConcurrencyLimitDataSourceModel struct {
	// The model requires the same fields, so reuse the fields defined in the resource model.
	resources.GlobalConcurrencyLimitResourceModel
}

// NewGlobalConcurrencyLimitDataSource returns a new GlobalConcurrencyLimitDataSource.
//
//nolint:ireturn // required by Terraform API
func NewGlobalConcurrencyLimitDataSource() datasource.DataSource {
	return &globalConcurrencyLimitDataSource{}
}

// Metadata returns the data source type name.
func (d *globalConcurrencyLimitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_concurrency_limit"
}

// Configure initializes runtime state for the data source.
func (d *globalConcurrencyLimitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *globalConcurrencyLimitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Global Concurrency Limit
<br>
Use this data source to read down the pre-defined Global Concurrency Limits, to manage concurrency limits.
<br>
For more information, see [apply global concurrency and rate limits](https://docs.prefect.io/v3/develop/global-concurrency-limits).
`,
			helpers.AllPlans...,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Global Concurrency Limit ID (UUID)",
				Optional:    true,
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
				Optional:    true,
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
				Optional:    true,
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the global concurrency limit",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum number of tasks that can run simultaneously",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the global concurrency limit is active.",
			},
			"active_slots": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of active slots.",
			},
			"slot_decay_per_second": schema.Float64Attribute{
				Computed:    true,
				Description: "The number of slots to decay per second.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *globalConcurrencyLimitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model globalConcurrencyLimitDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If both ID and Name are unset, we cannot read the global concurrency limit
	if model.ID.IsNull() && model.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"Either a Global Concurrency Limit ID or Name are required to read a Global Concurrency Limit.",
		)

		return
	}

	// Get the global concurrency limit
	client, err := d.client.GlobalConcurrencyLimits(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("global concurrency limit", err))

		return
	}

	// A global concurrency limit can be read by either ID or Name
	// If both are set, we prefer the ID
	var limit *api.GlobalConcurrencyLimit
	var operation string
	var getErr error

	switch {
	case !model.ID.IsNull():
		operation = "get by ID"
		limit, getErr = client.Read(ctx, model.ID.ValueString())
	case !model.Name.IsNull() && model.ID.IsNull():
		operation = "get by Name"
		limit, getErr = client.Read(ctx, model.Name.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Either id, or name are unset",
			"Please configure either id, or name.",
		)

		return
	}

	if getErr != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("global concurrency limit", operation, getErr))

		return
	}

	model.ID = customtypes.NewUUIDValue(limit.ID)
	model.Created = customtypes.NewTimestampPointerValue(limit.Created)
	model.Updated = customtypes.NewTimestampPointerValue(limit.Updated)
	model.AccountID = customtypes.NewUUIDValue(limit.AccountID)
	model.WorkspaceID = customtypes.NewUUIDValue(limit.WorkspaceID)

	model.Name = types.StringValue(limit.Name)
	model.Limit = types.Int64Value(limit.Limit)
	model.Active = types.BoolValue(limit.Active)
	model.ActiveSlots = types.Int64Value(limit.ActiveSlots)
	model.SlotDecayPerSecond = types.Float64Value(limit.SlotDecayPerSecond)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
