package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"k8s.io/utils/ptr"
)

var (
	_ = resource.ResourceWithConfigure(&WorkQueueResource{})
	_ = resource.ResourceWithImportState(&WorkQueueResource{})
)

// WorkQueueResource contains state for the resource.
type WorkQueueResource struct {
	client api.PrefectClient
}

// WorkQueueResourceModel defines the Terraform resource model.
type WorkQueueResourceModel struct {
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

// NewWorkQueueResource returns a new WorkQueueResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkQueueResource() resource.Resource {
	return &WorkQueueResource{}
}

// Metadata returns the resource type name.
func (r *WorkQueueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_queue"
}

// Configure initializes runtime state for the resource.
func (r *WorkQueueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *WorkQueueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `work_queue` represents a Prefect Work Queue. "+
				"Work Queues are used to configure and manage job execution queues in Prefect.\n"+
				"\n"+
				"Work Queues can be associated with a work pool and have configurations like concurrency limits.\n"+
				"\n"+
				"For more information, see [work queues](https://docs.prefect.io/v3/deploy/infrastructure-concepts/work-pools#work-queues).",
			helpers.AllPlans...,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Work queue ID (UUID)",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider. In Prefect Cloud, either the `work_pool` resource or the provider's `workspace_id` must be set.",
				Optional:    true,
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
			},
			"work_pool_name": schema.StringAttribute{
				Description: "The name of the work pool associated with this work queue",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the work queue",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the work queue",
				Default:     stringdefault.StaticString(""), // Because prefect returns this as the default for none provided
			},
			"is_paused": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether this work queue is paused",
				Optional:    true,
			},
			"concurrency_limit": schema.Int64Attribute{
				Description: "The concurrency limit applied to this work queue",
				Optional:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "The priority of this work queue",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// copyWorkQueueToModel maps an API response to a model that is saved in Terraform state.
func copyWorkQueueToModel(queue *api.WorkQueue, tfModel *WorkQueueResourceModel) {
	tfModel.ID = customtypes.NewUUIDValue(queue.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(queue.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(queue.Updated)
	tfModel.ConcurrencyLimit = types.Int64PointerValue(queue.ConcurrencyLimit)
	tfModel.Priority = types.Int64PointerValue(queue.Priority)
	tfModel.Description = types.StringPointerValue(queue.Description)
	tfModel.Name = types.StringValue(queue.Name)
	tfModel.IsPaused = types.BoolValue(queue.IsPaused)
	tfModel.WorkPoolName = types.StringValue(queue.WorkPoolName)
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkQueueResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkQueues(
		plan.AccountID.ValueUUID(),
		plan.WorkspaceID.ValueUUID(),
		plan.WorkPoolName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Queue", err))

		return
	}

	createRequest := api.WorkQueueCreate{
		Name:             plan.Name.ValueString(),
		IsPaused:         plan.IsPaused.ValueBoolPointer(),
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64Pointer(),
		Description:      plan.Description.ValueStringPointer(),
	}

	if *plan.Priority.ValueInt64Pointer() != 0 { // 0 is the value it provides when user does not specify it
		createRequest.Priority = plan.Priority.ValueInt64Pointer()
	}

	// Create the work queue using the WorkQueue client
	queue, err := client.Create(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "create", err))

		return
	}

	copyWorkQueueToModel(queue, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkQueueResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkQueues(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID(), state.WorkPoolName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Queue", err))

		return
	}

	queue, err := client.Get(ctx, state.Name.ValueString())
	if err != nil {
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		// Otherwise, we can log the error diagnostic and return
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "get", err))

		return
	}

	copyWorkQueueToModel(queue, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkQueueResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkQueues(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID(), plan.WorkPoolName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Queue", err))

		return
	}

	updateRequest := api.WorkQueueUpdate{
		IsPaused:         plan.IsPaused.ValueBoolPointer(),
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64Pointer(),
	}

	if plan.Description.IsNull() {
		updateRequest.Description = ptr.To("")
	} else {
		updateRequest.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Priority.IsNull() {
		updateRequest.Priority = plan.Priority.ValueInt64Pointer()
	}

	err = client.Update(ctx, plan.Name.ValueString(), updateRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "update", err))

		return
	}

	queue, err := client.Get(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "get", err))

		return
	}

	copyWorkQueueToModel(queue, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkQueueResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkQueues(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID(), state.WorkPoolName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Queue", err))

		return
	}

	err = client.Delete(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Queue", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Allow input values in the form of:
	// - "work_pool_name,work_queue_name"
	// - "work_pool_name,work_queue_name,workspace_id"
	minInputCount := 2
	maxInputCount := 3
	inputParts := strings.Split(req.ID, ",")

	switch inputCount := len(inputParts); inputCount {
	case minInputCount:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("work_pool_name"), inputParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), inputParts[1])...)
	case maxInputCount:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("work_pool_name"), inputParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), inputParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), inputParts[2])...)
	default:
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifiers, in the form of `work_pool_name,work_queue_name,workspace_id` or `work_pool_name,work_queue_name`. Got %q", req.ID),
		)

		return
	}
}
