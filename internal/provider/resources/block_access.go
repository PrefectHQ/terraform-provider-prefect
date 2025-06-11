package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

type BlockAccessResource struct {
	client api.PrefectClient
}

type BlockAccessResourceModel struct {
	BlockID        customtypes.UUIDValue `tfsdk:"block_id"`
	ManageActorIDs types.List            `tfsdk:"manage_actor_ids"`
	ViewActorIDs   types.List            `tfsdk:"view_actor_ids"`
	ManageTeamIDs  types.List            `tfsdk:"manage_team_ids"`
	ViewTeamIDs    types.List            `tfsdk:"view_team_ids"`
	AccountID      customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID    customtypes.UUIDValue `tfsdk:"workspace_id"`
}

// NewBlockAccessResource returns a new BlockAccessResource.
//
//nolint:ireturn // required by Terraform API
func NewBlockAccessResource() resource.Resource {
	return &BlockAccessResource{}
}

// Metadata returns the resource type name.
func (r *BlockAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_access"
}

// Configure initializes runtime state for the resource.
func (r *BlockAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BlockAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
This resource manages access control to Blocks. Accessors can be Service Accounts, Users, or Teams. Accessors can be Managers or Viewers.

All Actors/Teams must first be granted access to the Workspace where the Block resides.

Leave fields empty to use the default access controls
`,
			helpers.PlanPrefectCloudPro,
			helpers.PlanPrefectCloudEnterprise,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"block_id": schema.StringAttribute{
				Required:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Block ID (UUID)",
			},
			"manage_actor_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of actor IDs with manage access to the Block",
				ElementType: types.StringType,
			},
			"view_actor_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of actor IDs with view access to the Block",
				ElementType: types.StringType,
			},
			"manage_team_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of team IDs with manage access to the Block",
				ElementType: types.StringType,
			},
			"view_team_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyList),
				Description: "List of team IDs with view access to the Block",
				ElementType: types.StringType,
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the Block is located",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) where the Block is located. In Prefect Cloud, either the `prefect_block_access` resource or the provider's `workspace_id` must be set.",
			},
		},
	}
}
func (r *BlockAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	var manageActorIDs []string
	resp.Diagnostics.Append(plan.ManageActorIDs.ElementsAs(ctx, &manageActorIDs, false)...)
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

	var viewTeamIDs []string
	resp.Diagnostics.Append(plan.ViewTeamIDs.ElementsAs(ctx, &viewTeamIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := api.BlockDocumentAccessUpsert{}
	payload.AccessControl.ManageActorIDs = manageActorIDs
	payload.AccessControl.ViewActorIDs = viewActorIDs
	payload.AccessControl.ManageTeamIDs = manageTeamIDs
	payload.AccessControl.ViewTeamIDs = viewTeamIDs

	err = client.UpsertAccess(ctx, plan.BlockID.ValueUUID(), payload)

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "Upsert Access", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *BlockAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlockAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.BlockDocuments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	_, err = client.GetAccess(ctx, state.BlockID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "Get Access", err))

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

func (r *BlockAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BlockAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	var manageActorIDs []string
	resp.Diagnostics.Append(plan.ManageActorIDs.ElementsAs(ctx, &manageActorIDs, false)...)
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

	var viewTeamIDs []string
	resp.Diagnostics.Append(plan.ViewTeamIDs.ElementsAs(ctx, &viewTeamIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := api.BlockDocumentAccessUpsert{}
	payload.AccessControl.ManageActorIDs = manageActorIDs
	payload.AccessControl.ViewActorIDs = viewActorIDs
	payload.AccessControl.ManageTeamIDs = manageTeamIDs
	payload.AccessControl.ViewTeamIDs = viewTeamIDs
	err = client.UpsertAccess(ctx, plan.BlockID.ValueUUID(), payload)

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "Upsert Access", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *BlockAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlockAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.BlockDocuments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	payload := api.BlockDocumentAccessUpsert{}
	payload.AccessControl.ManageActorIDs = []string{}
	payload.AccessControl.ViewActorIDs = []string{}
	payload.AccessControl.ManageTeamIDs = []string{}
	payload.AccessControl.ViewTeamIDs = []string{}
	err = client.UpsertAccess(ctx, state.BlockID.ValueUUID(), payload)

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "Upsert Access", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
