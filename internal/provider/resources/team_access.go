package resources

import (
	"context"
	"fmt"

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
		Description: helpers.DescriptionWithPlans(
			"The resource `team_access` grants access to a team for a user or service account. "+
				"For more information, see [manage teams](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams).",
			helpers.PlanEnterprise,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Team Access ID",
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
				Validators: []validator.String{
					stringvalidator.OneOf("user", "service_account"),
				},
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// waitForTeamAccessToExist waits for the team access to be available after creation.
// This handles eventual consistency where the team access may be created but not
// immediately available for reading.
func waitForTeamAccessToExist(ctx context.Context, client api.TeamAccessClient, teamID, memberID, memberActorID customtypes.UUIDValue) (*api.TeamAccess, error) {
	teamAccess, err := helpers.WaitForResourceStabilization(
		ctx,
		func(ctx context.Context) (*api.TeamAccess, error) {
			return client.Read(ctx, teamID.ValueUUID(), memberID.ValueUUID(), memberActorID.ValueUUID())
		},
		func(_ *api.TeamAccess) error {
			// If we successfully read the team access, it exists
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for team access to exist: %w", err)
	}

	return teamAccess, nil
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

	// Wait for the team access to be available after creation
	teamAccess, err := waitForTeamAccessToExist(ctx, client, plan.TeamID, plan.MemberID, plan.MemberActorID)
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
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

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
