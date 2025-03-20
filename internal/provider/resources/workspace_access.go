package resources

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

var _ = resource.ResourceWithConfigure(&WorkspaceAccessResource{})

type WorkspaceAccessResource struct {
	client api.PrefectClient
}

type WorkspaceAccessResourceModel struct {
	ID              types.String          `tfsdk:"id"`
	AccessorType    types.String          `tfsdk:"accessor_type"`
	AccessorID      customtypes.UUIDValue `tfsdk:"accessor_id"`
	WorkspaceRoleID customtypes.UUIDValue `tfsdk:"workspace_role_id"`

	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
}

// NewWorkspaceAccessResource returns a new WorkspaceAccessResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceAccessResource() resource.Resource {
	return &WorkspaceAccessResource{}
}

// Metadata returns the resource type name.
func (r *WorkspaceAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_access"
}

// Configure initializes runtime state for the resource.
func (r *WorkspaceAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the schema for the WorkspaceAccessResource.
func (r *WorkspaceAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `workspace_access` represents a connection between an accessor "+
				"(User, Service Account or Team) with a Workspace Role. This resource specifies an actor's access level "+
				"to a specific Workspace in the Account.\n"+
				"\n"+
				"Use this resource in conjunction with the `workspace_role` resource or data source to manage access to Workspaces.\n"+
				"\n"+
				"For more information, see [manage workspaces](https://docs.prefect.io/v3/manage/cloud/workspaces).",
			helpers.PlanPrefectCloudPro,
			helpers.PlanPrefectCloudEnterprise,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace Access ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"accessor_type": schema.StringAttribute{
				Required:    true,
				Description: "USER | SERVICE_ACCOUNT | TEAM",
				Validators: []validator.String{
					stringvalidator.OneOf(utils.ServiceAccount, utils.User, utils.Team),
				},
			},
			"accessor_id": schema.StringAttribute{
				Required:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "ID (UUID) of accessor to the workspace. This can be an `account_member.user_id` or `service_account.id`",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the workspace is located",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) to grant access to",
			},
			"workspace_role_id": schema.StringAttribute{
				Required:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace Role ID (UUID) to grant to accessor",
			},
		},
	}
}

// copyWorkspaceAccessToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyWorkspaceAccessToModel(access *api.WorkspaceAccess, tfModel *WorkspaceAccessResourceModel) {
	// NOTE: api.WorkspaceAccess represents a combined model for all accessor types,
	// meaning accessor-specific attributes like BotID and UserID will be conditionally nil
	// depending on the accessor type.
	tfModel.ID = types.StringValue(access.ID.String())
	tfModel.WorkspaceRoleID = customtypes.NewUUIDValue(access.WorkspaceRoleID)
	tfModel.WorkspaceID = customtypes.NewUUIDValue(access.WorkspaceID)

	if access.BotID != nil {
		tfModel.AccessorID = customtypes.NewUUIDValue(*access.BotID)
	}
	if access.UserID != nil {
		tfModel.AccessorID = customtypes.NewUUIDValue(*access.UserID)
	}
	if access.TeamID != nil {
		tfModel.AccessorID = customtypes.NewUUIDValue(*access.TeamID)
	}
}

// Create will create the Workspace Access resource through the API and insert it into the State.
func (r *WorkspaceAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkspaceAccessResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceAccess(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace Access", err))

		return
	}

	accessorType := plan.AccessorType.ValueString()

	workspaceAccess, err := client.Upsert(ctx, accessorType, plan.AccessorID.ValueUUID(), plan.WorkspaceRoleID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Access", "create", err))

		return
	}

	copyWorkspaceAccessToModel(workspaceAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkspaceAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkspaceAccessResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceAccess(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace Access", err))

		return
	}

	accessID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace Role", err))

		return
	}

	accessorType := state.AccessorType.ValueString()

	workspaceAccess, err := client.Get(ctx, accessorType, accessID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Access", "read", err))

		return
	}

	copyWorkspaceAccessToModel(workspaceAccess, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkspaceAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkspaceAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkspaceAccess(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace Access", err))

		return
	}

	accessorType := plan.AccessorType.ValueString()

	workspaceAccess, err := client.Upsert(ctx, accessorType, plan.AccessorID.ValueUUID(), plan.WorkspaceRoleID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Access", "update", err))

		return
	}

	copyWorkspaceAccessToModel(workspaceAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkspaceAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkspaceAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, err := r.client.WorkspaceAccess(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace Access", err))

		return
	}

	accessID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace Access", err))

		return
	}

	accessorID := state.AccessorID.ValueUUID()
	accessorType := state.AccessorType.ValueString()

	err = client.Delete(ctx, accessorType, accessID, accessorID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Access", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
