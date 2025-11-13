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
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&WorkspaceResource{})
	_ = resource.ResourceWithImportState(&WorkspaceResource{})
)

// WorkspaceResource contains state for the resource.
type WorkspaceResource struct {
	client api.PrefectClient
}

// WorkspaceResourceModel defines the Terraform resource model.
type WorkspaceResourceModel struct {
	BaseModel

	AccountID customtypes.UUIDValue `tfsdk:"account_id"`

	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
}

// NewWorkspaceResource returns a new WorkspaceResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

// Metadata returns the resource type name.
func (r *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

// Configure initializes runtime state for the resource.
func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *WorkspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans("The resource `workspace` represents a Prefect Cloud Workspace. "+
			"Workspaces are discrete environments in Prefect Cloud for your flows, configurations, and deployments. "+
			"Manage your workflows and RBAC policies using `work_pool` and `workspace_access` resources. "+
			"For more information, see [manage workspaces](https://docs.prefect.io/v3/manage/cloud/workspaces).",
			helpers.AllCloudPlans...,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID)",
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
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

// waitForWorkspaceStateStabilization waits for the workspace's fields to match the expected values.
// This handles eventual consistency where the workspace may be created or updated with values
// that get updated asynchronously by the API.
func waitForWorkspaceStateStabilization(ctx context.Context, client api.WorkspacesClient, workspaceID uuid.UUID, expectedName, expectedHandle, expectedDescription string) (*api.Workspace, error) {
	workspace, err := helpers.WaitForResourceStabilization(
		ctx,
		func(ctx context.Context) (*api.Workspace, error) {
			return client.Get(ctx, workspaceID)
		},
		func(workspace *api.Workspace) error {
			// Check if all fields match expected values
			if workspace.Name != expectedName {
				return fmt.Errorf("workspace name does not match expected: got %q, want %q", workspace.Name, expectedName)
			}
			if workspace.Handle != expectedHandle {
				return fmt.Errorf("workspace handle does not match expected: got %q, want %q", workspace.Handle, expectedHandle)
			}
			// Description can be nil, so we need to handle that case
			actualDescription := ""
			if workspace.Description != nil {
				actualDescription = *workspace.Description
			}
			if actualDescription != expectedDescription {
				return fmt.Errorf("workspace description does not match expected: got %q, want %q", actualDescription, expectedDescription)
			}

			// All fields match, we're done
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for workspace state stabilization: %w", err)
	}

	return workspace, nil
}

// copyWorkspaceModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyWorkspaceToModel(_ context.Context, workspace *api.Workspace, tfModel *WorkspaceResourceModel) diag.Diagnostics {
	tfModel.ID = customtypes.NewUUIDValue(workspace.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(workspace.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(workspace.Updated)

	tfModel.Name = types.StringValue(workspace.Name)
	tfModel.Handle = types.StringValue(workspace.Handle)
	tfModel.Description = types.StringValue(*workspace.Description)

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Workspaces(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace", err))

		return
	}

	workspace, err := client.Create(
		ctx,
		api.WorkspaceCreate{
			Name:        plan.Name.ValueString(),
			Handle:      plan.Handle.ValueString(),
			Description: plan.Description.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", "create", err))

		return
	}

	// Wait for the workspace state to stabilize
	// The API may update fields asynchronously after creation
	expectedDescription := ""
	if plan.Description.ValueStringPointer() != nil {
		expectedDescription = *plan.Description.ValueStringPointer()
	}
	workspace, err = waitForWorkspaceStateStabilization(
		ctx,
		client,
		workspace.ID,
		plan.Name.ValueString(),
		plan.Handle.ValueString(),
		expectedDescription,
	)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", "create", err))

		return
	}

	resp.Diagnostics.Append(copyWorkspaceToModel(ctx, workspace, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkspaceResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && state.Handle.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Handle are unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	client, err := r.client.Workspaces(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace", err))

		return
	}

	// A workspace can be imported + read by either ID or Handle
	// If both are set, we prefer the ID
	var workspace *api.Workspace
	var operation string

	if !state.ID.IsNull() {
		var workspaceID uuid.UUID
		workspaceID, err = uuid.Parse(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

			return
		}

		operation = "get"
		workspace, err = client.Get(ctx, workspaceID)
	} else if !state.Handle.IsNull() {
		var workspaces []*api.Workspace
		operation = "list"
		workspaces, err = client.List(ctx, []string{state.Handle.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific workspace is not returned
		if err == nil && len(workspaces) != 1 {
			err = fmt.Errorf("a workspace with the handle=%s could not be found", state.Handle.ValueString())
		}

		if len(workspaces) == 1 {
			workspace = workspaces[0]
		}
	}

	if err != nil {
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError(
			"Error refreshing Workspace state",
			fmt.Sprintf("Could not read Workspace, unexpected error: %s", err.Error()),
		)
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", operation, err))

		return
	}

	resp.Diagnostics.Append(copyWorkspaceToModel(ctx, workspace, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Workspaces(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace", err))

		return
	}

	workspaceID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

		return
	}

	payload := api.WorkspaceUpdate{
		Name:        plan.Name.ValueStringPointer(),
		Handle:      plan.Handle.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
	}
	err = client.Update(ctx, workspaceID, payload)

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", "update", err))

		return
	}

	// Wait for the workspace state to stabilize after update
	// The API may update fields asynchronously
	expectedDescription := ""
	if plan.Description.ValueStringPointer() != nil {
		expectedDescription = *plan.Description.ValueStringPointer()
	}
	workspace, err := waitForWorkspaceStateStabilization(
		ctx,
		client,
		workspaceID,
		plan.Name.ValueString(),
		plan.Handle.ValueString(),
		expectedDescription,
	)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", "get", err))

		return
	}

	resp.Diagnostics.Append(copyWorkspaceToModel(ctx, workspace, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkspaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Workspaces(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace", err))

		return
	}

	workspaceID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

		return
	}

	err = client.Delete(ctx, workspaceID)
	if err != nil {
		// If the resource is already deleted (404), treat as success for idempotent deletion.
		// This can happen when cleanup processes or cascading deletes have already removed the workspace.
		if helpers.Is404Error(err) {
			return
		}

		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.HasPrefix(req.ID, "handle/") {
		handle := strings.TrimPrefix(req.ID, "handle/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("handle"), handle)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}
