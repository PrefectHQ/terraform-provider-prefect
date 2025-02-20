package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = datasource.DataSourceWithConfigure(&WorkQueuesDataSource{})

// WorkQueuesDataSource contains state for the data source.
type WorkQueuesDataSource struct {
	client api.PrefectClient
}

// WorkQueuesSourceModel defines the Terraform data source model.
type WorkQueuesSourceModel struct {
	AccountID    customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID  customtypes.UUIDValue `tfsdk:"workspace_id"`
	WorkPoolName types.String          `tfsdk:"work_pool_name"`

	FilterAny  types.List `tfsdk:"filter_any"`
	WorkQueues types.Set  `tfsdk:"work_queues"`
}

// NewWorkQueuesDataSource returns a new WorkQueuesDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkQueuesDataSource() datasource.DataSource {
	return &WorkQueuesDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkQueuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_queues"
}

// Configure initializes runtime state for the data source.
func (d *WorkQueuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *WorkQueuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about multiple Work Queues.
<br>
Use this data source to search for multiple Work Queues. Defaults to fetching all Work Queues in the Workspace.
<br>
For more information, see [work queues](https://docs.prefect.io/v3/deploy/infrastructure-concepts/work-pools#work-queues).
`,
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
				Description: "Work queue IDs (UUID) to search for (work queues with any matching UUID are returned)",
			},
			"work_queues": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Work queues returned by the server",
				NestedObject: schema.NestedAttributeObject{
					Attributes: workQueueAttributesBase,
				},
			},
			"work_pool_name": schema.StringAttribute{
				Description: "Name of the associated work pool",
				Required:    true, // Required for now because of my design of work queue client
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkQueuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkQueuesSourceModel

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

	filter := api.WorkQueueFilter{}

	// List work queues with the filter
	queues, err := client.List(ctx, filter)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queues", "list", err))

		return
	}

	attributeTypes := map[string]attr.Type{
		"id":                customtypes.UUIDType{},
		"created":           customtypes.TimestampType{},
		"updated":           customtypes.TimestampType{},
		"name":              types.StringType,
		"description":       types.StringType,
		"is_paused":         types.BoolType,
		"concurrency_limit": types.Int64Type,
		"priority":          types.Int64Type,
		"work_pool_name":    types.StringType,
	}

	// Map each work queue to its attributes
	queueObjects := make([]attr.Value, 0, len(queues))
	for _, queue := range queues {
		attributeValues := map[string]attr.Value{
			"id":                customtypes.NewUUIDValue(queue.ID),
			"created":           customtypes.NewTimestampPointerValue(queue.Created),
			"updated":           customtypes.NewTimestampPointerValue(queue.Updated),
			"name":              types.StringValue(queue.Name),
			"description":       types.StringPointerValue(queue.Description),
			"is_paused":         types.BoolValue(queue.IsPaused),
			"concurrency_limit": types.Int64PointerValue(queue.ConcurrencyLimit),
			"priority":          types.Int64PointerValue(queue.Priority),
			"work_pool_name":    types.StringValue(queue.WorkPoolName),
		}

		// Convert the attributes to match the expected type
		queueObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		queueObjects = append(queueObjects, queueObject)
	}

	// Set the final list value to be returned
	set, diag := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: attributeTypes}, queueObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.WorkQueues = set

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
