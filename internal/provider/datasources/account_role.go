package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &AccountRoleDataSource{}
var _ datasource.DataSourceWithConfigure = &AccountRoleDataSource{}

type AccountRoleDataSource struct {
	client api.PrefectClient
}

// AccountRoleDataSource defines the Terraform data source model
// the TF data source configuration will be unmarshalled into this struct.
type AccountRoleDataSourceModel struct {
	BaseModel

	Name         types.String          `tfsdk:"name"`
	Permissions  types.List            `tfsdk:"permissions"`
	AccountID    customtypes.UUIDValue `tfsdk:"account_id"`
	IsSystemRole types.Bool            `tfsdk:"is_system_role"`
}

// NewWorkspaceRoleDataSource returns a new WorkspaceRoleDataSource.
//
//nolint:ireturn // required by Terraform API
func NewAccountRoleDataSource() datasource.DataSource {
	return &AccountRoleDataSource{}
}

// Metadata returns the data source type name.
func (d *AccountRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_role"
}

// Schema defines the schema for the data source.
func (d *AccountRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Workspace Role.
<br>
Use this data source read down the pre-defined Roles, to manage User and Service Account access.
<br>
For more information, see [manage account roles](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-roles#manage-account-roles).
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account Role ID (UUID)",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Account Role",
				Validators: []validator.String{
					stringvalidator.OneOf("Admin", "Member", "Owner"),
				},
			},
			"permissions": schema.ListAttribute{
				Computed:    true,
				Description: "List of permissions linked to the Account Role",
				ElementType: types.StringType,
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the resource resides",
			},
			"is_system_role": schema.BoolAttribute{
				Computed:    true,
				Description: "Boolean specifying if the Account Role is a system role",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *AccountRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *AccountRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccountRoleDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.AccountRoles(config.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account Roles", err))

		return
	}

	// Fetch an existing Account Role by name
	// Here, we'd expect only 1 Role (or none) to be returned
	// as we are querying a single Role name, not a list of names
	accountRoles, err := client.List(ctx, []string{config.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account Role", "list", err))

		return
	}

	if len(accountRoles) != 1 {
		resp.Diagnostics.AddError(
			"Could not find Workspace Role",
			fmt.Sprintf("Could not find Workspace Role with name %s", config.Name.String()),
		)

		return
	}

	fetchedRole := accountRoles[0]

	config.ID = customtypes.NewUUIDValue(fetchedRole.ID)
	config.Created = customtypes.NewTimestampPointerValue(fetchedRole.Created)
	config.Updated = customtypes.NewTimestampPointerValue(fetchedRole.Updated)

	config.Name = types.StringValue(fetchedRole.Name)
	config.AccountID = customtypes.NewUUIDPointerValue(fetchedRole.AccountID)
	config.IsSystemRole = types.BoolValue(fetchedRole.IsSystemRole)

	list, diags := types.ListValueFrom(ctx, types.StringType, fetchedRole.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Permissions = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
