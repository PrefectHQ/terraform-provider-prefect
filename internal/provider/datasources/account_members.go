package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &AccountMembershipsDataSource{}
var _ datasource.DataSourceWithConfigure = &AccountMembershipsDataSource{}

type AccountMembershipsDataSource struct {
	client api.PrefectClient
}

type AccountMembershipsDataSourceModel struct {
	Members types.List `tfsdk:"members"`

	AccountID customtypes.UUIDValue `tfsdk:"account_id"`
}

// NewAccountMemberDataSource returns a new AccountMemberDataSource.
//
//nolint:ireturn // required by Terraform API
func NewAccountMembershipsDataSource() datasource.DataSource {
	return &AccountMembershipsDataSource{}
}

// Metadata returns the data source type name.
func (d *AccountMembershipsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_members"
}

// Schema defines the schema for the data source.
func (d *AccountMembershipsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about all members of account.
<br>
Use this data source to obtain user or actor IDs to manage Workspace Access.
`,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"members": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Account members of an account",
				NestedObject: schema.NestedAttributeObject{
					Attributes: accountMemberAttributesBase,
				},
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *AccountMembershipsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *AccountMembershipsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model AccountMembershipsDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.AccountMemberships(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account Memberships", err))

		return
	}

	// Fetch all existing account members
	var filter []string
	accountMembers, err := client.List(ctx, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Account Members state",
			fmt.Sprintf("Could not retrieve Account Members, unexpected error: %s", err.Error()),
		)
	}

	attributeTypes := map[string]attr.Type{
		"id":                customtypes.UUIDType{},
		"actor_id":          customtypes.UUIDType{},
		"user_id":           customtypes.UUIDType{},
		"first_name":        types.StringType,
		"last_name":         types.StringType,
		"handle":            types.StringType,
		"email":             types.StringType,
		"account_role_id":   customtypes.UUIDType{},
		"account_role_name": types.StringType,
	}

	memberObjects := make([]attr.Value, 0, len(accountMembers))

	for _, accountMember := range accountMembers {

		attributeValues := map[string]attr.Value{

			"id":                customtypes.NewUUIDValue(accountMember.ID),
			"actor_id":          customtypes.NewUUIDValue(accountMember.ActorID),
			"user_id":           customtypes.NewUUIDValue(accountMember.UserID),
			"first_name":        types.StringValue(accountMember.FirstName),
			"last_name":         types.StringValue(accountMember.LastName),
			"handle":            types.StringValue(accountMember.Handle),
			"email":             types.StringValue(accountMember.Email),
			"account_role_id":   customtypes.NewUUIDValue(accountMember.AccountRoleID),
			"account_role_name": types.StringValue(accountMember.AccountRoleName),
		}

		memberObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		memberObjects = append(memberObjects, memberObject)
	}

	list, diag := types.ListValue(types.ObjectType{AttrTypes: attributeTypes}, memberObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Members = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
