package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&DeploymentAccessResource{})

type DeploymentAccessResource struct {
	client api.PrefectClient
}

type DeploymentAccessResourceModel struct {
	ID            types.String          `tfsdk:"id"`
	DeploymentID  customtypes.UUIDValue `tfsdk:"deployment_id"`
	AccessControl types.Object          `tfsdk:"access_control"`

	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
}

func (m DeploymentAccessResourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"deployment_id":  types.StringType,
		"account_id":     types.StringType,
		"workspace_id":   types.StringType,
		"access_control": types.ObjectType{AttrTypes: DeploymentAccessControlResourceModel{}.AttrTypes()},
	}
}

type DeploymentAccessControlResourceModel struct {
	ManageActorIDs types.List `tfsdk:"manage_actor_ids"`
	RunActorIDs    types.List `tfsdk:"run_actor_ids"`
	ViewActorIDs   types.List `tfsdk:"view_actor_ids"`
	ManageTeamIDs  types.List `tfsdk:"manage_team_ids"`
	RunTeamIDs     types.List `tfsdk:"run_team_ids"`
	ViewTeamIDs    types.List `tfsdk:"view_team_ids"`
}

func (m DeploymentAccessControlResourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"manage_actor_ids": types.ListType{ElemType: types.StringType},
		"run_actor_ids":    types.ListType{ElemType: types.StringType},
		"view_actor_ids":   types.ListType{ElemType: types.StringType},
		"manage_team_ids":  types.ListType{ElemType: types.StringType},
		"run_team_ids":     types.ListType{ElemType: types.StringType},
		"view_team_ids":    types.ListType{ElemType: types.StringType},
	}
}

// NewDeploymentAccessResource returns a new DeploymentAccessResource.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentAccessResource() resource.Resource {
	return &DeploymentAccessResource{}
}

// Metadata returns the resource type name.
func (r *DeploymentAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_access"
}

// Configure initializes runtime state for the resource.
func (r *DeploymentAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description: "Resource representing Prefect Workspace Access for a User or Service Account",
		Description: "Set access controls for a deployment. " +
			"The given access controls will replace any existing access controls. ",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace Access ID (UUID)",
				// attributes which are not configurable + should not show updates from the existing state value
				// should implement `UseStateForUnknown()`
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_control": schema.SingleNestedAttribute{
				Description: "Access Control",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"manage_actor_ids": schema.ListAttribute{
						Description: "The list of actor IDs to grant manage access to.",
						ElementType: types.StringType,
						Required:    true,
					},
					"run_actor_ids": schema.ListAttribute{
						Description: "The list of actor IDs to grant run access to.",
						ElementType: types.StringType,
						Required:    true,
					},
					"view_actor_ids": schema.ListAttribute{
						Description: "The list of actor IDs to grant view access to.",
						ElementType: types.StringType,
						Required:    true,
					},
					"manage_team_ids": schema.ListAttribute{
						Description: "The list of team IDs to grant manage access to.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"run_team_ids": schema.ListAttribute{
						Description: "The list of team IDs to grant run access to.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"view_team_ids": schema.ListAttribute{
						Description: "The list of team IDs to grant view access to.",
						ElementType: types.StringType,
						Optional:    true,
					},
				},
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
			"deployment_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Deployment ID (UUID) to grant access to",
			},
		},
	}
}

// copyDeploymentAccessToModel copies the API resource to the Terraform model.
// Note that api.DeploymentAccess represents a combined model for all accessor types,
// meaning accessory-specific attributes like BotID and UserID will be conditionally nil
// depending on the accessor type.
func copyDeploymentAccessToModel(access *api.DeploymentAccess, model *DeploymentAccessResourceModel) {
	ctx := context.TODO()

	model.ID = types.StringValue(access.ID.String())
	model.DeploymentID = customtypes.NewUUIDValue(access.DeploymentID)
	model.WorkspaceID = customtypes.NewUUIDValue(access.WorkspaceID)
	model.AccountID = customtypes.NewUUIDValue(access.AccountID)

	model.AccessControl, _ = types.ObjectValueFrom(ctx, DeploymentAccessResourceModel{}.AttrTypes(), access.AccessControl)
}

// Create will create the Workspace Access resource through the API and insert it into the State.
func (r *DeploymentAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentAccessResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(data.AccountID.ValueUUID(), data.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployments", err))
		return
	}

	accessControlReq := api.DeploymentAccessSet{
		DeploymentID: data.DeploymentID.ValueUUID(),
		// AccessControl: api.DeploymentAccessControl{},
	}

	var accessControl DeploymentAccessControlResourceModel
	resp.Diagnostics.Append(data.AccessControl.As(ctx, &accessControl, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessControlReq.AccessControl.ManageActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}
	accessControlReq.AccessControl.ViewActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}

	if !accessControl.ManageActorIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.ManageActorIDs.ElementsAs(ctx, &accessControlReq.AccessControl.ManageActorIDs, false)...)
	}
	if !accessControl.RunActorIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.RunActorIDs.ElementsAs(ctx, &accessControlReq.AccessControl.RunActorIDs, false)...)
	}
	if !accessControl.ViewActorIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.ViewActorIDs.ElementsAs(ctx, &accessControlReq.AccessControl.ViewActorIDs, false)...)
	}
	if !accessControl.ManageTeamIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.ManageTeamIDs.ElementsAs(ctx, &accessControlReq.AccessControl.ManageTeamIDs, false)...)
	}
	if !accessControl.RunTeamIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.RunTeamIDs.ElementsAs(ctx, &accessControlReq.AccessControl.RunTeamIDs, false)...)
	}
	if !accessControl.ViewTeamIDs.IsNull() {
		resp.Diagnostics.Append(accessControl.ViewTeamIDs.ElementsAs(ctx, &accessControlReq.AccessControl.ViewTeamIDs, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentAccess, err := client.SetAccess(ctx, accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Set", err))
		return
	}

	copyDeploymentAccessToModel(deploymentAccess, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *DeploymentAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentAccessResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(data.AccountID.ValueUUID(), data.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployments", err))

		return
	}

	// accessID, err := uuid.Parse(data.ID.ValueString())
	// if err != nil {
	// 	resp.Diagnostics.AddAttributeError(
	// 		path.Root("id"),
	// 		"Error parsing Deployment Access ID",
	// 		fmt.Sprintf("Could not parse Deployment Access ID to UUID, unexpected error: %s", err.Error()),
	// 	)
	// }

	accessControlReq := api.DeploymentAccessRead{
		DeploymentID: data.DeploymentID.ValueUUID(),
	}

	deploymentAccess, err := client.ReadAccess(ctx, accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Read", err))

		return
	}

	copyDeploymentAccessToModel(deploymentAccess, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DeploymentAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeploymentAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployments", err))
		return
	}

	accessControlReq := api.DeploymentAccessSet{
		DeploymentID: plan.DeploymentID.ValueUUID(),
	}
	plan.AccessControl.As(ctx, plan.AccessControl, basetypes.ObjectAsOptions{})

	deploymentAccess, err := client.SetAccess(ctx, accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Set", err))

		return
	}

	copyDeploymentAccessToModel(deploymentAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DeploymentAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace Access ID",
			fmt.Sprintf("Could not parse Workspace Access ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	accessorType := state.AccessorType.ValueString()

	err = client.Delete(ctx, accessorType, accessID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Access", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
