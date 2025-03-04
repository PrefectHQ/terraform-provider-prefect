package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&TeamAccessResource{})

// TeamAccessResource is a resource that manages the members of a team.
type TeamAccessResource struct {
	client api.PrefectClient
}

// TeamAccessResourceModel is the model for the TeamAccessResource.
type TeamAccessResourceModel struct {
	ID customtypes.UUIDValue `tfsdk:"id"`

	TeamID        customtypes.UUIDValue `tfsdk:"team_id"`
	MemberID      customtypes.UUIDValue `tfsdk:"member_id"`
	MemberActorID customtypes.UUIDValue `tfsdk:"member_actor_id"`
	MemberType    types.String          `tfsdk:"member_type"`

	AccountID customtypes.UUIDValue `tfsdk:"account_id"`
}

// NewTeamAccessResource returns a new TeamAccessResource.
//
//nolint:ireturn // required by Terraform API
func NewTeamAccessResource() resource.Resource {
	return &TeamAccessResource{}
}

// Metadata returns the resource type name.
func (r *TeamAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_access"
}

// Configure initializes runtime state for the resource.
func (r *TeamAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the schema for the TeamAccessResource.
func (r *TeamAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `team` represents a Prefect Team. " +
			"Teams are used to organize users and their permissions. " +
			"For more information, see [manage teams](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Team Access ID",
				CustomType:  customtypes.UUIDType{},
			},
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "Team ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"member_id": schema.StringAttribute{
				Required:    true,
				Description: "Member ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"member_actor_id": schema.StringAttribute{
				Required:    true,
				Description: "Member Actor ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"member_type": schema.StringAttribute{
				Required:    true,
				Description: "Member Type (user | service_account)",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID)",
			},
		},
	}
}

// Create creates a new team access.
func (r *TeamAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TeamAccess(plan.AccountID.ValueUUID(), plan.TeamID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Team Access", err))

		return
	}

	if err := client.Upsert(ctx, plan.MemberType.ValueString(), plan.MemberID.ValueUUID()); err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "create", err))

		return
	}

	teamAccess, err := client.Read(ctx, plan.TeamID.ValueUUID(), plan.MemberID.ValueUUID(), plan.MemberActorID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "create", err))

		return
	}

	copyTeamAccessToModel(teamAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the team access.
func (r *TeamAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan TeamAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TeamAccess(plan.AccountID.ValueUUID(), plan.TeamID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Team Access", err))

		return
	}

	teamAccess, err := client.Read(ctx, plan.TeamID.ValueUUID(), plan.MemberID.ValueUUID(), plan.MemberActorID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "read", err))

		return
	}

	copyTeamAccessToModel(teamAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the team access.
func (r *TeamAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamAccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TeamAccess(plan.AccountID.ValueUUID(), plan.TeamID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Team Access", err))

		return
	}

	if err := client.Upsert(ctx, plan.MemberType.ValueString(), plan.MemberID.ValueUUID()); err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "update", err))
	}

	teamAccess, err := client.Read(ctx, plan.TeamID.ValueUUID(), plan.MemberID.ValueUUID(), plan.MemberActorID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "update", err))

		return
	}

	copyTeamAccessToModel(teamAccess, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the team access.
func (r *TeamAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan TeamAccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.TeamAccess(plan.AccountID.ValueUUID(), plan.TeamID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Team Access", err))

		return
	}

	err = client.Delete(ctx, plan.MemberActorID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team Access", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func copyTeamAccessToModel(teamAccess *api.TeamAccess, plan *TeamAccessResourceModel) {
	plan.ID = customtypes.NewUUIDValue(teamAccess.ID)
	plan.TeamID = customtypes.NewUUIDValue(teamAccess.TeamID)
	plan.MemberID = customtypes.NewUUIDValue(teamAccess.MemberID)
	plan.MemberActorID = customtypes.NewUUIDValue(teamAccess.MemberActorID)
	plan.MemberType = types.StringValue(teamAccess.MemberType)
}
