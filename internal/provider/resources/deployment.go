package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var (
	_ = resource.ResourceWithConfigure(&DeploymentResource{})
	_ = resource.ResourceWithImportState(&DeploymentResource{})
)

// DeploymentResource contains state for the resource.
type DeploymentResource struct {
	client api.PrefectClient
}

// DeploymentResourceModel defines the Terraform resource model.
type DeploymentResourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`

	Name        types.String          `tfsdk:"name"`
	Version     types.String          `tfsdk:"version"`
	Description types.String          `tfsdk:"description"`
	FlowID      customtypes.UUIDValue `tfsdk:"flow_id"`

	// schedule
	// IntervalSchedule (object) or CronSchedule (object) or RRuleSchedule (object) (Schedule)
	// A schedule for the deployment.

	IsScheduleActive types.Bool `tfsdk:"is_schedule_active"`
	Paused           types.Bool `tfsdk:"paused"`

	// schedules
	// Array of objects (Schedules)
	// A list of schedules for the deployment.

	// 	job_variables
	// object (Job Variables)
	// Overrides to apply to the base infrastructure block at runtime.

	// parameters
	// object (Parameters)
	// Parameters for flow runs scheduled by the deployment.

	Tags          types.List   `tfsdk:"tags"`
	WorkQueueName types.String `tfsdk:"work_queue_name"`
	// LastPolled    customtypes.TimestampValue `tfsdk:"last_polled"`

	// parameter_openapi_schema
	// object (Parameter Openapi Schema)
	// The parameter schema of the flow, including defaults.

	Path types.String `tfsdk:"path"`

	// pull_steps
	// Array of objects (Pull Steps)
	// The steps required to pull this deployment's project to a remote location.

	Entrypoint               types.String `tfsdk:"entrypoint"`
	ManifestPath             types.String `tfsdk:"manifest_path"`
	StorageDocumentID        types.String `tfsdk:"storage_document_id"`
	InfrastructureDocumentID types.String `tfsdk:"infrastructure_document_id"`
	WorkPoolName             types.String `tfsdk:"work_pool_name"`
	EnforceParameterSchema   types.Bool   `tfsdk:"enforce_parameter_schema"`
}

// NewDeploymentResource returns a new DeploymentResource.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// Metadata returns the resource type name.
func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

// Configure initializes runtime state for the resource.
func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		// Description: "Resource representing a Prefect Workspace",
		Description: "Deployments are server-side representations of flows. " +
			"They store the crucial metadata needed for remote orchestration including when, where, and how a workflow should run. " +
			"Deployments elevate workflows from functions that you must call manually to API-managed entities that can be triggered remotely.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				// We cannot use a CustomType due to a conflict with PlanModifiers; see
				// https://github.com/hashicorp/terraform-plugin-framework/issues/763
				// https://github.com/hashicorp/terraform-plugin-framework/issues/754
				Description: "Workspace ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) to associate deployment to",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the workspace",
				Required:    true,
			},
			"flow_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Flow ID (UUID) to associate deployment to",
				Required:    true,
			},
			"is_schedule_active": schema.BoolAttribute{
				Description: "Is Schedule Active",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"paused": schema.BoolAttribute{
				Description: "Whether or not the deployment is paused.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// schedules
			// Array of objects (Schedules)
			// A list of schedules for the deployment.
			"enforce_parameter_schema": schema.BoolAttribute{
				Description: "Whether or not the deployment should enforce the parameter schema.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// parameter_openapi_schema
			// object (Parameter Openapi Schema)
			// The parameter schema of the flow, including defaults.
			// parameters
			// object (Parameters)
			// Parameters for flow runs scheduled by the deployment.
			// pull_steps
			// Array of objects (Pull Steps)
			// The steps required to pull this deployment's project to a remote location.
			"manifest_path": schema.StringAttribute{
				Description: "The path to the flow's manifest file, relative to the chosen storage.",
				Computed:    true,
			},
			"work_queue_name": schema.StringAttribute{
				Description: "The work queue for the deployment. If no work queue is set, work will not be scheduled.",
				Optional:    true,
			},
			"work_pool_name": schema.StringAttribute{
				Description: "The name of the deployment's work pool.",
				Optional:    true,
			},
			"storage_document_id": schema.StringAttribute{
				Description: "The block document defining storage used for this flow.",
				Optional:    true,
			},
			"infrastructure_document_id": schema.StringAttribute{
				Description: "The block document defining infrastructure to use for flow runs.",
				Optional:    true,
			},
			// schedule
			// IntervalSchedule (object) or CronSchedule (object) or RRuleSchedule (object) (Schedule)
			// A schedule for the deployment.
			"description": schema.StringAttribute{
				Description: "A description for the deployment.",
				Optional:    true,
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path to the working directory for the workflow, relative to remote storage or an absolute path.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "An optional version for the deployment.",
				Optional:    true,
			},
			"entrypoint": schema.StringAttribute{
				Description: "The path to the entrypoint for the workflow, relative to the path.",
				Optional:    true,
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the deployment",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyTagList),
			},
		},
	}
}

// copyDeploymentToModel copies an api.Deployment to a DeploymentResourceModel.
func copyDeploymentToModel(ctx context.Context, deployment *api.Deployment, model *DeploymentResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(deployment.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(deployment.Created)
	model.Updated = customtypes.NewTimestampPointerValue(deployment.Updated)

	model.Name = types.StringValue(deployment.Name)
	// model.WorkspaceID = customtypes.NewUUIDValue(deployment.WorkspaceID)
	model.FlowID = customtypes.NewUUIDValue(deployment.FlowID)
	model.IsScheduleActive = types.BoolValue(deployment.IsScheduleActive)
	model.Paused = types.BoolValue(deployment.Paused)
	model.EnforceParameterSchema = types.BoolValue(deployment.EnforceParameterSchema)
	// model.ManifestPath = types.StringValue(deployment.ManifestPath)
	// model.WorkQueueName = types.StringValue(deployment.WorkQueueName)
	// model.WorkPoolName = types.StringValue(deployment.WorkPoolName)
	// model.StorageDocumentID = types.StringValue(deployment.StorageDocumentID)
	// model.InfrastructureDocumentID = types.StringValue(deployment.InfrastructureDocumentID)
	model.Description = types.StringValue(deployment.Description)
	model.Path = types.StringValue(deployment.Path)
	// model.Version = types.StringValue(deployment.Version)
	model.Entrypoint = types.StringValue(deployment.Entrypoint)

	tags, diags := types.ListValueFrom(ctx, types.StringType, deployment.Tags)
	if diags.HasError() {
		return diags
	}
	model.Tags = tags

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeploymentResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := client.Create(ctx, api.DeploymentCreate{
		Name:                   plan.Name.ValueString(),
		FlowID:                 plan.FlowID.ValueUUID(),
		IsScheduleActive:       plan.IsScheduleActive.ValueBool(),
		Paused:                 plan.Paused.ValueBool(),
		EnforceParameterSchema: plan.EnforceParameterSchema.ValueBool(),
		Path:                   plan.Path.ValueString(),
		Entrypoint:             plan.Entrypoint.ValueString(),
		Description:            plan.Description.ValueString(),
		Tags:                   tags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment",
			fmt.Sprintf("Could not create deployment, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model DeploymentResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	// A deployment can be imported + read by either ID or Handle
	// If both are set, we prefer the ID
	var deployment *api.Deployment
	if !model.ID.IsNull() {
		var deploymentID uuid.UUID
		deploymentID, err = uuid.Parse(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error parsing Deployment ID",
				fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}

		deployment, err = client.Get(ctx, deploymentID)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing deployment state",
			fmt.Sprintf("Could not read Deployment, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(model.WorkspaceID.ValueUUID(), model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	deploymentID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Deployment ID",
			fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	payload := api.DeploymentUpdate{
		Name: model.Name.ValueStringPointer(),
		// Handle:      model.Handle.ValueStringPointer(),
		// Description: model.Description.ValueStringPointer(),
	}
	err = client.Update(ctx, deploymentID, payload)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating deployment",
			fmt.Sprintf("Could not update deployment, unexpected error: %s", err),
		)

		return
	}

	deployment, err := client.Get(ctx, deploymentID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Deployment state",
			fmt.Sprintf("Could not read Deployment, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	deploymentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Deployment ID",
			fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, deploymentID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Deployment",
			fmt.Sprintf("Could not delete Deployment, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// we'll allow input values in the form of:
	// - "id,workspace_id"
	// - "id"
	maxInputCount := 2
	inputParts := strings.Split(req.ID, ",")

	// eg. "foo,bar,baz"
	if len(inputParts) > maxInputCount {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	// eg. ",foo" or "foo,"
	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected non-empty import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	identifier := inputParts[0]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), identifier)...)

	if len(inputParts) == 2 && inputParts[1] != "" {
		workspaceID, err := uuid.Parse(inputParts[1])
		if err != nil {
			resp.Diagnostics.AddError("problem parsing workspace ID", "see details")

			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID.String())...)
	}
}

// // ImportState imports the resource into Terraform state.
// func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	// we'll allow input values in the form of:
// 	// - "workspace_id,name"
// 	// - "name"
// 	maxInputCount := 2
// 	inputParts := strings.Split(req.ID, ",")

// 	// eg. "foo,bar,baz"
// 	if len(inputParts) > maxInputCount {
// 		resp.Diagnostics.AddError(
// 			"Unexpected Import Identifier",
// 			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `workspace_id,name`. Got %q", req.ID),
// 		)

// 		return
// 	}

// 	// eg. ",foo" or "foo,"
// 	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
// 		resp.Diagnostics.AddError(
// 			"Unexpected Import Identifier",
// 			fmt.Sprintf("Expected non-empty import identifiers, in the form of `workspace_id,name`. Got %q", req.ID),
// 		)

// 		return
// 	}

// 	if len(inputParts) == maxInputCount {
// 		workspaceID := inputParts[0]
// 		id := inputParts[1]

// 		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
// 		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
// 	} else {
// 		resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
// 	}
// }
