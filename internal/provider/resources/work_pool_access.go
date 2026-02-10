package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&WorkPoolAccessResource{})

type WorkPoolAccessResource struct {
	client api.PrefectClient
}

type WorkPoolAccessResourceModel struct {
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	WorkPoolName types.String `tfsdk:"work_pool_name"`

	ManageActorIDs types.List `tfsdk:"manage_actor_ids"`
	RunActorIDs    types.List `tfsdk:"run_actor_ids"`
	ViewActorIDs   types.List `tfsdk:"view_actor_ids"`
	ManageTeamIDs  types.List `tfsdk:"manage_team_ids"`
	RunTeamIDs     types.List `tfsdk:"run_team_ids"`
	ViewTeamIDs    types.List `tfsdk:"view_team_ids"`
}

// NewWorkPoolAccessResource returns a new WorkPoolAccessResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkPoolAccessResource() resource.Resource {
	return &WorkPoolAccessResource{}
}

// Metadata returns the resource type name.
func (r *WorkPoolAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pool_access"
}

// Configure initializes runtime state for the resource.
func (r *WorkPoolAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the resource schema.
func (r *WorkPoolAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `work_pool_access` represents a connection between an accessor "+
				"(User, Service Account or Team) with a Work Pool. This resource specifies an actor's access level "+
				"to a specific Work Pool in the Account. "+
				"For more information, see [object access control lists](https://docs.prefect.io/v3/manage/cloud/manage-users/object-access-control-lists).",
			helpers.PlanPro,
			helpers.PlanEnterprise,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Account ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "Workspace ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"work_pool_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Work Pool",
			},
			"manage_actor_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				ElementType: types.StringType,
				Description: "List of actor IDs with manage access to the Work Pool",
			},
			"run_actor_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				ElementType: types.StringType,
				Description: "List of actor IDs with run access to the Work Pool",
			},
			"view_actor_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				ElementType: types.StringType,
				Description: "List of actor IDs with view access to the Work Pool",
			},
			"manage_team_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of team IDs with manage access to the Work Pool",
				ElementType: types.StringType,
			},
			"run_team_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of team IDs with run access to the Work Pool",
				ElementType: types.StringType,
			},
			"view_team_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of team IDs with view access to the Work Pool",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkPoolAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkPoolAccessResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPoolAccess(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool Access", err))

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

	err = client.Set(ctx, plan.WorkPoolName.ValueString(), api.WorkPoolAccessSet{
		AccessControl: api.WorkPoolAccessControlSet{
			ManageActorIDs: manageActorIDs,
			ManageTeamIDs:  manageTeamIDs,
			RunActorIDs:    runActorIDs,
			RunTeamIDs:     runTeamIDs,
			ViewActorIDs:   viewActorIDs,
			ViewTeamIDs:    viewTeamIDs,
		},
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool Access", "create", err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource and sets the Terraform state.
func (r *WorkPoolAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkPoolAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPoolAccess(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool Access", err))

		return
	}

	_, err = client.Read(ctx, state.WorkPoolName.ValueString())
	if err != nil {
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool Access", "read", err))

		return
	}

	// NOTE: we are not currently mapping the response back into State,
	// as the Read payload is materially different from the Create/Update payloads.
	// This is something to be revisited in the future.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the Terraform state.
func (r *WorkPoolAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkPoolAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPoolAccess(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool Access", err))

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

	err = client.Set(ctx, plan.WorkPoolName.ValueString(), api.WorkPoolAccessSet{
		AccessControl: api.WorkPoolAccessControlSet{
			ManageActorIDs: manageActorIDs,
			ManageTeamIDs:  manageTeamIDs,
			RunActorIDs:    runActorIDs,
			RunTeamIDs:     runTeamIDs,
			ViewActorIDs:   viewActorIDs,
			ViewTeamIDs:    viewTeamIDs,
		},
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool Access", "update", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resets the access control to empty.
func (r *WorkPoolAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkPoolAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPoolAccess(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool Access", err))

		return
	}

	payload := api.WorkPoolAccessSet{}
	payload.AccessControl.ManageActorIDs = []string{}
	payload.AccessControl.ViewActorIDs = []string{}
	payload.AccessControl.ManageTeamIDs = []string{}
	payload.AccessControl.ViewTeamIDs = []string{}
	payload.AccessControl.RunActorIDs = []string{}
	payload.AccessControl.RunTeamIDs = []string{}

	err = client.Set(ctx, state.WorkPoolName.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool Access", "delete", err))

		return
	}
}
