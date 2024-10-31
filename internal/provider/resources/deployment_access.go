package resources

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&DeploymentAccessResource{})

type DeploymentAccessResource struct {
	client api.PrefectClient
}

type DeploymentAccessResourceModel struct {
	ID types.String `tfsdk:"id"`

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	DeploymentID customtypes.UUIDValue `tfsdk:"deployment_id"`

	ManageActorIDs types.List `tfsdk:"manage_actor_ids"`
	RunActorIDs    types.List `tfsdk:"run_actor_ids"`
	ViewActorIDs   types.List `tfsdk:"view_actor_ids"`
	ManageTeamIDs  types.List `tfsdk:"manage_team_ids"`
	RunTeamIDs     types.List `tfsdk:"run_team_ids"`
	ViewTeamIDs    types.List `tfsdk:"view_team_ids"`
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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

func (r *DeploymentAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `deployment_access` represents a connection between an accessor " +
			"(User, Service Account or Team) with a Deployment. This resource specifies an actor's access level " +
			"to a specific Deployment in the Account.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment Access ID (UUID)",
			},
			"deployment_id": schema.StringAttribute{
				Required:    true,
				Description: "Deployment ID (UUID)",
				CustomType:  customtypes.UUIDType{},
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
			"manage_actor_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Manage Actor IDs",
			},
			"run_actor_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Run Actor IDs",
			},
			"view_actor_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "View Actor IDs",
			},
			"manage_team_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Manage Team IDs",
			},
			"run_team_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Run Team IDs",
			},
			"view_team_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "View Team IDs",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *DeploymentAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeploymentAccessResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.DeploymentAccess(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Access", err))

		return
	}

	var manageActorIDs []string
	resp.Diagnostics.Append(plan.ManageActorIDs.ElementsAs(ctx, &manageActorIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var runActorIDs []string
	resp.Diagnostics.Append(plan.RunActorIDs.ElementsAs(ctx, &runActorIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var viewActorIDs []string
	resp.Diagnostics.Append(plan.ViewActorIDs.ElementsAs(ctx, &viewActorIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var manageTeamIDs []string
	resp.Diagnostics.Append(plan.ManageTeamIDs.ElementsAs(ctx, &manageTeamIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var runTeamIDs []string
	resp.Diagnostics.Append(plan.RunTeamIDs.ElementsAs(ctx, &runTeamIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var viewTeamIDs []string
	resp.Diagnostics.Append(plan.ViewTeamIDs.ElementsAs(ctx, &viewTeamIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = client.Set(ctx, plan.DeploymentID.ValueUUID(), api.DeploymentAccessSet{
		AccessControl: api.DeploymentAccessControlSet{
			ManageActorIDs: manageActorIDs,
			RunActorIDs:    runActorIDs,
			ViewActorIDs:   viewActorIDs,
			ManageTeamIDs:  manageTeamIDs,
			RunTeamIDs:     runTeamIDs,
			ViewTeamIDs:    viewTeamIDs,
		},
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "create", err))

		return
	}
}

// Read reads the resource and sets the Terraform state.
func (r *DeploymentAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeploymentAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Println("state", state)

	client, err := r.client.DeploymentAccess(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Access", err))

		return
	}

	_, err = client.Read(ctx, state.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Access", "read", err))

		return
	}

	// NOTE: we are not currently mapping the response back into State,
	// as the Read payload is materially different from the Set payload.
	// This is something to be revisited in the future.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the Terraform state.
func (r *DeploymentAccessResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// no-op, we don't update deployment access, only set it.
}

// Delete deletes the resource and removes the Terraform state.
func (r *DeploymentAccessResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// no-op, we don't delete deployment access, only update it.
}
