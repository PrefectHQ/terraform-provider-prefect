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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
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
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *WorkspaceRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource representing a Prefect Workspace Role",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace Role UUID",
				// attributes which are not configurable + should not show updates from the existing state value
				// should implement `UseStateForUnknown()`
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Date and time of the Workspace Role creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Date and time that the Workspace Role was last updated in RFC 3339 format",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Workspace Role",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the Workspace Role",
			},
			"scopes": schema.ListAttribute{
				Description: "List of scopes linked to the Workspace Role",
				ElementType: types.StringType,
				Optional:    true,
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account UUID, defaults to the account set in the provider",
				Computed:    true,
			},
			"inherited_role_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace Role UUID, whose permissions are inherited by this Workspace Role",
				Optional:    true,
			},
		},
	}
}

// copyWorkspaceRoleToModel copies an api.WorkspaceRole to a WorkspaceRoleDataSourceModel.
func copyWorkspaceRoleToModel(ctx context.Context, role *api.WorkspaceRole, model *WorkspaceRoleResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(role.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(role.Created)
	model.Updated = customtypes.NewTimestampPointerValue(role.Updated)

	model.Name = types.StringValue(role.Name)
	model.Description = types.StringPointerValue(role.Description)
	model.AccountID = customtypes.NewUUIDPointerValue(role.AccountID)
	model.InheritedRoleID = customtypes.NewUUIDPointerValue(role.InheritedRoleID)

	// NOTE: here, we'll omit updating the TF state with the scopes returned from the API
	// as scopes in Prefect Cloud have a hierarchical structure. This means that children scopes
	// can be returned based on what a practioner configures in TF / HCL, which will cause
	// conflicts on apply, as the retrieved state from the API will vary slightly with
	// the Terraform configuration. Therefore, the state will hold the user-defined scope parameters,
	// which will include any children scopes on the Prefect Cloud side.

	// scopes, diags := types.ListValueFrom(ctx, types.StringType, role.Scopes)
	// if diags.HasError() {
	// 	return diags
	// }
	// model.Scopes = scopes

	return nil
}

func (r *WorkspaceRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model WorkspaceRoleResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	resp.Diagnostics.Append(model.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	role, err := client.Create(ctx, api.WorkspaceRoleUpsert{
		Name:            model.Name.ValueString(),
		Description:     model.Description.ValueStringPointer(),
		Scopes:          scopes,
		InheritedRoleID: model.InheritedRoleID.ValueUUIDPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role",
			fmt.Sprintf("Could not create Workspace Role, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkspaceRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model WorkspaceRoleResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
	client, err := r.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	roleID, err := uuid.Parse(model.ID.ValueString())
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

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkspaceRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model WorkspaceRoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	var scopes []string
	resp.Diagnostics.Append(model.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace Role ID",
			fmt.Sprintf("Could not parse Workspace Role ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Update(ctx, roleID, api.WorkspaceRoleUpsert{
		Name:            model.Name.ValueString(),
		Description:     model.Description.ValueStringPointer(),
		Scopes:          scopes,
		InheritedRoleID: model.InheritedRoleID.ValueUUIDPointer(),
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

	resp.Diagnostics.Append(copyWorkspaceRoleToModel(ctx, role, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkspaceRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model WorkspaceRoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Workspace Role client",
			fmt.Sprintf("Could not create Workspace Role client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	roleID, err := uuid.Parse(model.ID.ValueString())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState allows Terraform to start managing a Workspace Role resource
func (r *WorkspaceRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
