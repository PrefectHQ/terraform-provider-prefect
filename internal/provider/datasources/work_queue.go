package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// WorkQueueDataSource contains state for the data source.
type WorkQueueDataSource struct {
	client api.PrefectClient
}

// WorkQueueDataSourceModel defines the Terraform data source model.
type WorkQueueDataSourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	IsPaused         types.Bool   `tfsdk:"is_paused"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
	Priority         types.Int64  `tfsdk:"priority"`
	WorkPoolName     types.String `tfsdk:"work_pool_name"`
}

// NewWorkQueueDataSource returns a new WorkQueueDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkQueueDataSource() datasource.DataSource {
	return &WorkQueueDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_queue"
}

// Configure initializes runtime state for the data source.
func (d *WorkQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Shared set of schema attributes between work_queue (singular)
// and work_queues (plural) datasources. Any work_queue (singular)
// specific attributes will be added to a deep copy in the Schema method.
var workQueueAttributesBase = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Work pool ID (UUID)",
		Optional:    true,
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the work queue creation in RFC 3339 format",
	},
	"priority": schema.Int64Attribute{
		Computed:    true,
		Description: "Priority of the work queue",
	},
	"work_pool_name": schema.StringAttribute{
		Optional:    true,
		Description: "Name of the associated work pool",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the work pool was last updated in RFC 3339 format",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name of the work queue",
		Optional:    true,
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of the work queue",
		Optional:    true,
	},
	"is_paused": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether this work queue is paused",
	},
	"concurrency_limit": schema.Int64Attribute{
		Computed:    true,
		Description: "The concurrency limit applied to this work queue",
		Optional:    true,
	},
}

// Schema defines the schema for the data source.
func (d *WorkQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Create a copy of the base attributes
	// and add the account/workspace ID overrides here
	// as they are not needed in the work_queues (plural) list

	workQueueAttributes := make(map[string]schema.Attribute)
	for k, v := range workQueueAttributesBase {
		workQueueAttributes[k] = v
	}

	workQueueAttributes["account_id"] = schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID), defaults to the account set in the provider",
		Optional:    true,
	}
	workQueueAttributes["workspace_id"] = schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
		Optional:    true,
	}

	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Work Queue by name.
<br>
Use this data source to obtain Work Queue-specific attributes.
`,
		Attributes: workQueueAttributes,
	}
}

// Read refreshes the Terraform state with the latest data for a Work Queue.
func (d *WorkQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkQueueDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.WorkQueues(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID(), model.WorkPoolName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Queue", err))

		return
	}

	queue, err := client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "get", err))

		return
	}

	// Map fetched data into the model
	model.ID = customtypes.NewUUIDValue(queue.ID)
	model.Created = customtypes.NewTimestampPointerValue(queue.Created)
	model.Updated = customtypes.NewTimestampPointerValue(queue.Updated)
	model.Description = types.StringPointerValue(queue.Description)
	model.IsPaused = types.BoolValue(queue.IsPaused)
	model.ConcurrencyLimit = types.Int64PointerValue(queue.ConcurrencyLimit)
	model.Priority = types.Int64PointerValue(queue.Priority)
	model.WorkPoolName = types.StringValue(queue.WorkPoolName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
