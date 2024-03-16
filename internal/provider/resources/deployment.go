package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
}

// NewDeploymentResource returns a new DeploymentResource.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// Metadata returns the resource type name.
func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

// Configure initializes runtime state for the resource.
func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Schema defines the schema for the resource.
func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description: "Resource representing a Prefect Workspace",
		Description: "The resource `workspace` represents a Prefect Cloud Workspace. " +
			"Workspaces are discrete environments in Prefect Cloud for your flows, configurations, and deployments. " +
			"Manage your workflows and RBAC policies using `work_pool` and `workspace_access` resources.",
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
			"name": schema.StringAttribute{
				Description: "Name of the workspace",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "Unique handle for the workspace",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for the workspace",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// copyDeploymentToModel copies an api.Workspace to a WorkspaceResourceModel.
func copyDeploymentToModel(_ context.Context, workspace *api.Deployment, model *DeploymentResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(workspace.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(workspace.Created)
	model.Updated = customtypes.NewTimestampPointerValue(workspace.Updated)

	model.Name = types.StringValue(workspace.Name)
	model.Handle = types.StringValue(workspace.Handle)
	model.Description = types.StringValue(*workspace.Description)

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model DeploymentResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(model.WorkspaceID.ValueUUID(), model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	deployment, err := client.Create(ctx, api.DeploymentCreate{
		Name:        model.Name.ValueString(),
		Handle:      model.Handle.ValueString(),
		Description: model.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace",
			fmt.Sprintf("Could not create workspace, unexpected error: %s", err),
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

// Read refreshes the Terraform state with the latest data.
func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model DeploymentResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() && model.Handle.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Handle are unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	client, err := r.client.Deployments(model.WorkspaceID.ValueUUID(), model.AccountID.ValueUUID())
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
	} else if !model.Handle.IsNull() {
		var deployments []*api.Deployment
		deployments, err = client.List(ctx, []string{model.Handle.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific deployment is not returned
		if err == nil && len(deployments) != 1 {
			err = fmt.Errorf("a deployment with the handle=%s could not be found", model.Handle.ValueString())
		}

		if len(deployments) == 1 {
			deployment = deployments[0]
		}
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
		Name:        model.Name.ValueStringPointer(),
		Handle:      model.Handle.ValueStringPointer(),
		Description: model.Description.ValueStringPointer(),
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
	var model DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Workspaces(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	workspaceID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace ID",
			fmt.Sprintf("Could not parse workspace ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Workspace",
			fmt.Sprintf("Could not delete Workspace, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.HasPrefix(req.ID, "handle/") {
		handle := strings.TrimPrefix(req.ID, "handle/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("handle"), handle)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}
