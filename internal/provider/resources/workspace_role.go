package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&WorkspaceRoleResource{})
	_ = resource.ResourceWithImportState(&WorkspaceRoleResource{})
)

// WorkspaceRoleResource contains state for the resource.
type WorkspaceRoleResource struct {
	client api.PrefectClient
}

// WorkspaceRoleResourceModel defines the Terraform resource model.
type WorkspaceRoleResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name            types.String          `tfsdk:"name"`
	Description     types.String          `tfsdk:"description"`
	Scopes          types.List            `tfsdk:"scopes"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	InheritedRoleID customtypes.UUIDValue `tfsdk:"inherited_role_id"`
}

// NewWorkspaceRoleResource returns a new WorkspaceRoleResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceRoleResource() resource.Resource {
	return &WorkspaceRoleResource{}
}

// Metadata returns the resource type name.
func (r *WorkspaceRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_role"
}

// Configure initializes runtime state for the resource.
func (r *WorkspaceRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkspaceRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description: "Resource representing a Prefect Workspace Role",
		Description: "The resource `workspace_role` represents a Prefect Cloud Workspace Role. " +
			"Workspace Roles hold a set of permissions to a specific Workspace, and can be attached to " +
			"an accessor (User or Service Account) to grant access to the Workspace.\n" +
			"\n" +
			"To obtain a list of available scopes, please refer to the `GET /api/workspace_scopes` " +
			"[API](https://app.prefect.cloud/api/docs#tag/Workspace-Scopes/operation/get_workspace_scopes_api_workspace_scopes_get)",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace Role ID (UUID)",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Workspace Role",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the Workspace Role",
				Default:     stringdefault.StaticString(""),
			},
			"scopes": schema.ListAttribute{
				Description: "List of scopes linked to the Workspace Role",
				ElementType: types.StringType,
				Optional:    true,
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Computed:    true,
			},
			"inherited_role_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace Role ID (UUID), whose permissions are inherited by this Workspace Role",
				Optional:    true,
			},
		},
	}
}

// copyWorkspaceRoleToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyWorkspaceRoleToModel(_ context.Context, role *api.WorkspaceRole, tfModel *WorkspaceRoleResourceModel) diag.Diagnostics {
	tfModel.ID = types.StringValue(role.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(role.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(role.Updated)

	tfModel.Name = types.StringValue(role.Name)
	tfModel.Description = types.StringPointerValue(role.Description)
	tfModel.AccountID = customtypes.NewUUIDPointerValue(role.AccountID)
	tfModel.InheritedRoleID = customtypes.NewUUIDPointerValue(role.InheritedRoleID)

	// NOTE: here, we'll omit updating the TF state with the scopes returned from the API
	// as scopes in Prefect Cloud have a hierarchical structure. For example, setting `manage_blocks`
	// will grant the subordinate `see_blocks` and `write_blocks` on the API side.  This means that
	// children scopes like this will be returned if you configure a `manage_*` scope, which will cause
	// conflicts on apply, as the retrieved state from the API will vary slightly with
	// the Terraform configuration. Therefore, the state will hold the user-defined scope parameters,
	// which will include any children scopes on the Prefect Cloud side.

	//nolint:gocritic
	// scopes, diags := types.ListValueFrom(ctx, types.StringType, role.Scopes)
	// if diags.HasError() {
	// 	return diags
	// }
	// model.Scopes = scopes

	return nil
}

func (r *WorkspaceRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceRoleResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	role, err := client.Create(ctx, api.WorkspaceRoleUpsert{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		Scopes:          scopes,
		InheritedRoleID: plan.InheritedRoleID.ValueUUIDPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role",
			fmt.Sprintf("Could not create Workspace Role, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkspaceRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkspaceRoleResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	roleID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace Role ID",
			fmt.Sprintf("Could not parse Workspace Role ID to UUID, unexpected error: %s", err.Error()),
		)
	}

	role, err := client.Get(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Workspace Role state",
			fmt.Sprintf("Could not read Workspace Role, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkspaceRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceRoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace Role ID",
			fmt.Sprintf("Could not parse Workspace Role ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Update(ctx, roleID, api.WorkspaceRoleUpsert{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		Scopes:          scopes,
		InheritedRoleID: plan.InheritedRoleID.ValueUUIDPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Workspace Role",
			fmt.Sprintf("Could not update Workspace Role, unexpected error: %s", err),
		)

		return
	}

	role, err := client.Get(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Workspace Role state",
			fmt.Sprintf("Could not read Workspace Role, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkspaceRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkspaceRoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	roleID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace Role ID",
			fmt.Sprintf("Could not parse Workspace Role ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Workspace Role",
			fmt.Sprintf("Could not delete Workspace Role, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState allows Terraform to start managing a Workspace Role resource.
func (r *WorkspaceRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
