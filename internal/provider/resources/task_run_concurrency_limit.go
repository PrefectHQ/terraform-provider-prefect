package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&TaskRunConcurrencyLimitResource{})
	_ = resource.ResourceWithImportState(&TaskRunConcurrencyLimitResource{})
)

// TaskRunConcurrencyLimitResource contains state for the resource.
type TaskRunConcurrencyLimitResource struct {
	client api.PrefectClient
}

// TaskRunConcurrencyLimitResourceModel defines the Terraform resource model.
type TaskRunConcurrencyLimitResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Tag              types.String `tfsdk:"tag"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
}

// NewTaskRunConcurrencyLimitResource returns a new TaskRunConcurrencyLimitResource.
//
//nolint:ireturn // required by Terraform API
func NewTaskRunConcurrencyLimitResource() resource.Resource {
	return &TaskRunConcurrencyLimitResource{}
}

// Metadata returns the resource type name.
func (r *TaskRunConcurrencyLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_run_concurrency_limit"
}

// Configure initializes runtime state for the resource.
func (r *TaskRunConcurrencyLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *TaskRunConcurrencyLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `task_run_concurrency_limit` represents a task run concurrency limit. Task run concurrency limits allow you to control how many tasks with specific tags can run simultaneously. For more information, see [limit concurrent task runs with tags](https://docs.prefect.io/v3/develop/task-run-limits).",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Task run concurrency limit ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Account ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "Workspace ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"tag": schema.StringAttribute{
				Required:    true,
				Description: "A tag the task run concurrency limit is applied to.",
				PlanModifiers: []planmodifier.String{
					// Task Run Concurrency limit updates are not supported so any changes to the tag will
					// require a replacement of the resource.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"concurrency_limit": schema.Int64Attribute{
				Required:    true,
				Description: "The task run concurrency limit.",
				PlanModifiers: []planmodifier.Int64{
					// Task Run Concurrency limit updates are not supported so any changes to the concurrency limit will
					// require a replacement of the resource.
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *TaskRunConcurrencyLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TaskRunConcurrencyLimitResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TaskRunConcurrencyLimits(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Task Run Concurrency Limit", err))

		return
	}

	taskRunConcurrencyLimit, err := client.Create(ctx, api.TaskRunConcurrencyLimitCreate{
		Tag:              plan.Tag.ValueString(),
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Task Run Concurrency Limit", "create", err))

		return
	}

	copyTaskRunConcurrencyLimitToModel(taskRunConcurrencyLimit, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func copyTaskRunConcurrencyLimitToModel(concurrencyLimit *api.TaskRunConcurrencyLimit, model *TaskRunConcurrencyLimitResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(concurrencyLimit.ID.String())
	model.Created = customtypes.NewTimestampValue(*concurrencyLimit.Created)
	model.Updated = customtypes.NewTimestampValue(*concurrencyLimit.Updated)
	model.Tag = types.StringValue(concurrencyLimit.Tag)
	model.ConcurrencyLimit = types.Int64Value(concurrencyLimit.ConcurrencyLimit)

	return nil
}

// Delete deletes the resource.
func (r *TaskRunConcurrencyLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TaskRunConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TaskRunConcurrencyLimits(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Task Run Concurrency Limit", err))

		return
	}

	err = client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Task Run Concurrency Limit", "delete", err))

		return
	}
}

// Read reads the resource state from the API.
func (r *TaskRunConcurrencyLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TaskRunConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TaskRunConcurrencyLimits(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Task Run Concurrency Limit", err))

		return
	}

	taskRunConcurrencyLimit, err := client.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Task Run Concurrency Limit", "get", err))

		return
	}

	copyTaskRunConcurrencyLimitToModel(taskRunConcurrencyLimit, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource state.
// This resource does not support updates.
func (r *TaskRunConcurrencyLimitResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// ImportState imports the resource into Terraform state.
func (r *TaskRunConcurrencyLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
