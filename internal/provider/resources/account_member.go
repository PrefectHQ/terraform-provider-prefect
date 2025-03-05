package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&AccountMemberResource{})
	_ = resource.ResourceWithImportState(&AccountMemberResource{})
)

// AccountMemberResource contains state for the resource.
type AccountMemberResource struct {
	client api.PrefectClient
}

// AccountMemberResourceModel defines the Terraform resource model.
type AccountMemberResourceModel struct {
	// This has the same fields as the AccountMemberDataSourceModel.
	ID              customtypes.UUIDValue `tfsdk:"id"`
	ActorID         customtypes.UUIDValue `tfsdk:"actor_id"`
	UserID          customtypes.UUIDValue `tfsdk:"user_id"`
	FirstName       types.String          `tfsdk:"first_name"`
	LastName        types.String          `tfsdk:"last_name"`
	Handle          types.String          `tfsdk:"handle"`
	Email           types.String          `tfsdk:"email"`
	AccountRoleID   customtypes.UUIDValue `tfsdk:"account_role_id"`
	AccountRoleName types.String          `tfsdk:"account_role_name"`

	AccountID customtypes.UUIDValue `tfsdk:"account_id"`
}

// NewAccountMemberResource returns a new AccountMemberResource.
//
//nolint:ireturn // required by Terraform API
func NewAccountMemberResource() resource.Resource {
	return &AccountMemberResource{}
}

// Metadata returns the resource type name.
func (r *AccountMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_member"
}

// Configure initializes runtime state for the resource.
func (r *AccountMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema defines the schema for the resource.
func (r *AccountMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `account_member` represents a member of an account. " +
			"It is used to manage the member's attributes, such as the actor_id, account_id, and account_role_id. " +
			"For more information, see [manage users](https://docs.prefect.io/v3/manage/cloud/manage-users",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account Member ID (UUID)",
			},
			"actor_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Actor ID (UUID)",
			},
			"user_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "User ID (UUID)",
			},
			"first_name": schema.StringAttribute{
				Computed:    true,
				Description: "Member's first name",
			},
			"last_name": schema.StringAttribute{
				Computed:    true,
				Description: "Member's last name",
			},
			"handle": schema.StringAttribute{
				Computed:    true,
				Description: "Member handle, or a human-readable identifier",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Member email",
			},
			"account_role_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Acount Role ID (UUID)",
			},
			"account_role_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of Account Role assigned to member",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID)",
			},
		},
	}
}

func copyAccountMemberToModel(_ context.Context, member *api.AccountMembership, tfModel *AccountMemberResourceModel) diag.Diagnostics {
	tfModel.ID = customtypes.NewUUIDValue(member.ID)
	tfModel.ActorID = customtypes.NewUUIDValue(member.ActorID)
	tfModel.UserID = customtypes.NewUUIDValue(member.UserID)
	tfModel.FirstName = types.StringValue(member.FirstName)
	tfModel.LastName = types.StringValue(member.LastName)
	tfModel.Handle = types.StringValue(member.Handle)
	tfModel.Email = types.StringValue(member.Email)
	tfModel.AccountRoleID = customtypes.NewUUIDValue(member.AccountRoleID)
	tfModel.AccountRoleName = types.StringValue(member.AccountRoleName)

	return nil
}

// Create creates the resource.
func (r *AccountMemberResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Not implemented", "Account member creation is not yet supported")
}

// Read reads the resource.
func (r *AccountMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(state.AccountID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.AccountMemberships(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	accountMembers, err := client.List(ctx, []string{state.Email.ValueString()})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "get", err))

		return
	}

	if len(accountMembers) != 1 {
		resp.Diagnostics.AddError(
			"Could not find Account Member",
			fmt.Sprintf("Could not find Account Member with email %s", state.Email.ValueString()),
		)

		return
	}

	fetchedAccountMember := accountMembers[0]

	resp.Diagnostics.Append(copyAccountMemberToModel(ctx, fetchedAccountMember, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource.
func (r *AccountMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state AccountMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(state.AccountID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.AccountMemberships(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	payload := api.AccountMembershipUpdate{
		AccountRoleID: state.AccountRoleID.ValueUUID(),
	}

	err = client.Update(ctx, state.ID.ValueUUID(), &payload)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "update", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource.
func (r *AccountMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(state.AccountID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.AccountMemberships(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	err = client.Delete(ctx, state.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "delete", err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
// Import syntax: <account_id>,email/<account_email>.
func (r *AccountMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	requiredParts := 2
	parts := strings.Split(req.ID, ",")

	if len(parts) != requiredParts {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format of <account_id>,email/<account_email>",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), parts[0])...)

	email := strings.TrimPrefix(parts[1], "email/")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), email)...)
}
