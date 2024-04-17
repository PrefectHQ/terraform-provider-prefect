package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	Result        types.Object          `tfsdk:"result"`

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
		"result":         types.ObjectType{AttrTypes: DeploymentAccessAccessResourceModel{}.AttrTypes()},
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

type DeploymentAccessAccessResourceModel struct {
	ManageActors types.List `tfsdk:"manage_actors"`
	RunActors    types.List `tfsdk:"run_actors"`
	ViewActors   types.List `tfsdk:"view_actors"`
}

func (m DeploymentAccessAccessResourceModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"manage_actors": types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentAccessControlActorModel{}.AttrTypes()}},
		"run_actors":    types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentAccessControlActorModel{}.AttrTypes()}},
		"view_actors":   types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentAccessControlActorModel{}.AttrTypes()}},
	}
}

type DeploymentAccessControlActorModel struct {
	ID    customtypes.UUIDValue `tfsdk:"id"`
	Name  types.String          `tfsdk:"name"`
	Email types.String          `tfsdk:"email"`
	Type  types.String          `tfsdk:"type"`
}

func (m DeploymentAccessControlActorModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":    customtypes.UUIDType{},
		"name":  types.StringType,
		"email": types.StringType,
		"type":  types.StringType,
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
						Optional:    true,
					},
					"run_actor_ids": schema.ListAttribute{
						Description: "The list of actor IDs to grant run access to.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"view_actor_ids": schema.ListAttribute{
						Description: "The list of actor IDs to grant view access to.",
						ElementType: types.StringType,
						Optional:    true,
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
			"result": schema.SingleNestedAttribute{
				Description: "Effective Access Control",
				Computed:    true,
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"manage_actors": schema.ListNestedAttribute{
						Description: "The list of actors granted access to manage.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									CustomType:  customtypes.UUIDType{},
									Description: "ID (UUID)",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name",
									Optional:    true,
								},
								"email": schema.StringAttribute{
									Computed:    true,
									Description: "Email",
									Optional:    true,
								},
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "Type",
									Optional:    true,
								},
							},
						},
					},
					"run_actors": schema.ListNestedAttribute{
						Description: "The list of actors granted access to run.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									CustomType:  customtypes.UUIDType{},
									Description: "ID (UUID)",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name",
									Optional:    true,
								},
								"email": schema.StringAttribute{
									Computed:    true,
									Description: "Email",
									Optional:    true,
								},
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "Type",
									Optional:    true,
								},
							},
						},
					},
					"view_actors": schema.ListNestedAttribute{
						Description: "The list of actors granted access to view.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									CustomType:  customtypes.UUIDType{},
									Description: "ID (UUID)",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name",
									Optional:    true,
								},
								"email": schema.StringAttribute{
									Computed:    true,
									Description: "Email",
									Optional:    true,
								},
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "Type",
									Optional:    true,
								},
							},
						},
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

func NewActorsModel(ctx context.Context, actors []api.Actor) (basetypes.ListValue, diag.Diagnostics) {
	var model = make([]DeploymentAccessControlActorModel, len(actors))
	for i, a := range actors {
		actor := DeploymentAccessControlActorModel{
			ID:    customtypes.NewUUIDValue(a.ID),
			Name:  types.StringValue(a.Name),
			Email: types.StringValue(a.Email),
			Type:  types.StringValue(a.Type),
		}
		model[i] = actor
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DeploymentAccessControlActorModel{}.AttrTypes()}, model)
}

// copyDeploymentAccessToModel copies the API resource to the Terraform model.
// Note that api.DeploymentAccess represents a combined model for all accessor types,
// meaning accessory-specific attributes like BotID and UserID will be conditionally nil
// depending on the accessor type.
func copyDeploymentAccessAccessToModel(ctx context.Context, access *api.DeploymentAccessControl, model *DeploymentAccessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	manageActors, diag := NewActorsModel(ctx, access.ManageActors)
	diags.Append(diag...)
	if diags.HasError() {
		return diags
	}
	runActors, diag := NewActorsModel(ctx, access.RunActors)
	diags.Append(diag...)
	if diags.HasError() {
		return diags
	}
	viewActors, diag := NewActorsModel(ctx, access.ViewActors)
	diags.Append(diag...)
	if diags.HasError() {
		return diags
	}

	result := DeploymentAccessAccessResourceModel{
		ManageActors: manageActors,
		RunActors:    runActors,
		ViewActors:   viewActors,
	}

	model.Result, diags = types.ObjectValueFrom(ctx, DeploymentAccessAccessResourceModel{}.AttrTypes(), result)
	return diags
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
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployments Access", err))
		return
	}

	accessControlReq := api.DeploymentAccessSet{}
	var accessControl DeploymentAccessControlResourceModel
	resp.Diagnostics.Append(data.AccessControl.As(ctx, &accessControl, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessControlReq.AccessControl.ManageActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}
	accessControlReq.AccessControl.ViewActorIDs = []string{}
	accessControlReq.AccessControl.ManageTeamIDs = []string{}
	accessControlReq.AccessControl.RunTeamIDs = []string{}
	accessControlReq.AccessControl.ViewTeamIDs = []string{}

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

	err = client.SetAccess(ctx, data.DeploymentID.ValueUUID(), accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Set", err))
		return
	}

	// Set the ID to the deployment ID value
	data.ID = data.DeploymentID.StringValue

	deploymentAccess, err := client.ReadAccess(ctx, data.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Read", err))
		return
	}
	resp.Diagnostics.Append(copyDeploymentAccessAccessToModel(ctx, deploymentAccess, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	deploymentAccess, err := client.ReadAccess(ctx, data.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Read", err))
		return
	}
	resp.Diagnostics.Append(copyDeploymentAccessAccessToModel(ctx, deploymentAccess, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	accessControlReq := api.DeploymentAccessSet{}
	plan.AccessControl.As(ctx, plan.AccessControl, basetypes.ObjectAsOptions{})

	err = client.SetAccess(ctx, plan.DeploymentID.ValueUUID(), accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Set", err))
		return
	}

	deploymentAccess, err := client.ReadAccess(ctx, plan.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Read", err))
		return
	}
	resp.Diagnostics.Append(copyDeploymentAccessAccessToModel(ctx, deploymentAccess, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DeploymentAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeploymentAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployments", err))
		return
	}

	accessControlReq := api.DeploymentAccessSet{}
	var accessControl DeploymentAccessControlResourceModel
	resp.Diagnostics.Append(state.AccessControl.As(ctx, &accessControl, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessControlReq.AccessControl.ManageActorIDs = []string{}
	accessControlReq.AccessControl.RunActorIDs = []string{}
	accessControlReq.AccessControl.ViewActorIDs = []string{}
	accessControlReq.AccessControl.ManageTeamIDs = []string{}
	accessControlReq.AccessControl.RunTeamIDs = []string{}
	accessControlReq.AccessControl.ViewTeamIDs = []string{}

	err = client.SetAccess(ctx, state.DeploymentID.ValueUUID(), accessControlReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Set", err))
		return
	}

	deploymentAccess, err := client.ReadAccess(ctx, state.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "Read", err))
		return
	}
	resp.Diagnostics.Append(copyDeploymentAccessAccessToModel(ctx, deploymentAccess, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
