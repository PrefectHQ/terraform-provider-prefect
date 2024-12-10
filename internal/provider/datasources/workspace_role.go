package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &WorkspaceRoleDataSource{}
var _ datasource.DataSourceWithConfigure = &WorkspaceRoleDataSource{}

type WorkspaceRoleDataSource struct {
	client api.PrefectClient
}

// WorkspaceRoleDataSourceModel defines the Terraform data source model
// the TF data source configuration will be unmarshalled into this struct.
type WorkspaceRoleDataSourceModel struct {
	BaseModel

	Name            types.String          `tfsdk:"name"`
	Description     types.String          `tfsdk:"description"`
	Scopes          types.List            `tfsdk:"scopes"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	InheritedRoleID customtypes.UUIDValue `tfsdk:"inherited_role_id"`
}

// NewWorkspaceRoleDataSource returns a new WorkspaceRoleDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceRoleDataSource() datasource.DataSource {
	return &WorkspaceRoleDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkspaceRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_role"
}

var workspaceRoleAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace Role ID (UUID)",
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
		Description: "Name of the Workspace Role",
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of the Workspace Role",
	},
	"scopes": schema.ListAttribute{
		Computed:    true,
		Description: "List of scopes linked to the Workspace Role",
		ElementType: types.StringType,
	},
	"account_id": schema.StringAttribute{
		Optional:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID) where Workspace Role resides",
	},
	"inherited_role_id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace Role ID (UUID), whose permissions are inherited by this Workspace Role",
	},
}

// Schema defines the schema for the data source.
func (d *WorkspaceRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Workspace Role.
<br>
Use this data source read down the pre-defined Roles, to manage User and Service Account access.
`,
		Attributes: workspaceRoleAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *WorkspaceRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *WorkspaceRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkspaceRoleDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Workspace Role", err))

		return
	}

	// Fetch an existing Workspace Role by name
	// Here, we'd expect only 1 Role (or none) to be returned
	// as we are querying a single Role name, not a list of names
	workspaceRoles, err := client.List(ctx, []string{model.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Workspace Role", "list", err))

		return
	}

	if len(workspaceRoles) != 1 {
		resp.Diagnostics.AddError(
			"Could not find Workspace Role",
			fmt.Sprintf("Could not find Workspace Role with name %s", model.Name.String()),
		)

		return
	}

	fetchedRole := workspaceRoles[0]

	model.ID = customtypes.NewUUIDValue(fetchedRole.ID)
	model.Created = customtypes.NewTimestampPointerValue(fetchedRole.Created)
	model.Updated = customtypes.NewTimestampPointerValue(fetchedRole.Updated)

	model.Name = types.StringValue(fetchedRole.Name)
	model.Description = types.StringPointerValue(fetchedRole.Description)
	model.AccountID = customtypes.NewUUIDPointerValue(fetchedRole.AccountID)
	model.InheritedRoleID = customtypes.NewUUIDPointerValue(fetchedRole.InheritedRoleID)

	list, diags := types.ListValueFrom(ctx, types.StringType, fetchedRole.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Scopes = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
