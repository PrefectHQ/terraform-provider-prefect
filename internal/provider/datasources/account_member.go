package datasources

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &AccountMemberDataSource{}
var _ datasource.DataSourceWithConfigure = &AccountMemberDataSource{}

type AccountMemberDataSource struct {
	client api.PrefectClient
}

type AccountMemberDataSourceModel struct {
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

// NewAccountMemberDataSource returns a new AccountMemberDataSource.
//
//nolint:ireturn // required by Terraform API
func NewAccountMemberDataSource() datasource.DataSource {
	return &AccountMemberDataSource{}
}

// Metadata returns the data source type name.
func (d *AccountMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_member"
}

// Shared set of schema attributes between account_member (singular)
// and account_members (plural) datasources. Any account_member (singular)
// specific attributes will be added to a deep copy in the Schema method.
var accountMemberAttributesBase = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Account Member ID (UUID)",
	},
	"actor_id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Actor ID (UUID), used for granting access to resources like Blocks and Deployments",
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
		Required:    true,
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
}

// Schema defines the schema for the data source.
func (d *AccountMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Create a copy of the base attributes
	// and add the account ID overrides here
	accountMemberAttributes := make(map[string]schema.Attribute)
	maps.Copy(accountMemberAttributes, accountMemberAttributesBase)
	accountMemberAttributes["account_id"] = schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID) where the member resides",
		Optional:    true,
	}

	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Account Member (user)	by their email.
<br>
Use this data source to obtain user or actor IDs to manage Workspace Access.
<br>
For more information, see [manage account roles](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams).
`,
			helpers.AllCloudPlans...,
		),
		Attributes: accountMemberAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *AccountMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *AccountMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccountMemberDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.AccountMemberships(config.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account Memberships", err))

		return
	}

	// Fetch an existing Account Member by email
	// Here, we'd expect only 1 Member (or none) to be returned
	// as we are querying a single Member email, not a list of emails
	// workspaceRoles, err := client.List(ctx, []string{model.Name.ValueString()})
	accountMembers, err := client.List(ctx, []string{config.Email.ValueString()})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account Member", "list", err))

		return
	}

	if len(accountMembers) != 1 {
		resp.Diagnostics.AddError(
			"Could not find Account Member",
			fmt.Sprintf("Could not find Account Member with email %s", config.Email.ValueString()),
		)

		return
	}

	fetchedAccountMember := accountMembers[0]

	config.ID = customtypes.NewUUIDValue(fetchedAccountMember.ID)
	config.ActorID = customtypes.NewUUIDValue(fetchedAccountMember.ActorID)
	config.UserID = customtypes.NewUUIDValue(fetchedAccountMember.UserID)
	config.FirstName = types.StringValue(fetchedAccountMember.FirstName)
	config.LastName = types.StringValue(fetchedAccountMember.LastName)
	config.Handle = types.StringValue(fetchedAccountMember.Handle)
	config.Email = types.StringValue(fetchedAccountMember.Email)
	config.AccountRoleID = customtypes.NewUUIDValue(fetchedAccountMember.AccountRoleID)
	config.AccountRoleName = types.StringValue(fetchedAccountMember.AccountRoleName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
