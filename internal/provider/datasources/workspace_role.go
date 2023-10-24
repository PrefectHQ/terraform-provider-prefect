package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
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
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name            types.String          `tfsdk:"name"`
	Description     types.String          `tfsdk:"description"`
	Permissions     types.List            `tfsdk:"permissions"`
	Scopes          types.List            `tfsdk:"scopes"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	InheritedRoleID customtypes.UUIDValue `tfsdk:"inherited_role_id"`
}

// NewWorkspaceRoleDataSource returns a new WorkspaceRoleDataSource.
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
		Description: "Workspace Role UUID",
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the Workspace Role creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the Workspace Role was last updated in RFC 3339 format",
	},
	"name": schema.StringAttribute{
		Required:    true,
		Description: "Name of the Workspace Role",
	},
	"description": schema.StringAttribute{
		Optional:    true,
		Description: "Description of the Workspace Role",
	},
	"permissions": schema.ListAttribute{
		Computed:    true,
		Description: "List of permissions linked to the Workspace Role",
		ElementType: types.StringType,
	},
	"scopes": schema.ListAttribute{
		Computed:    true,
		Description: "List of scopes linked to the Workspace Role",
		ElementType: types.StringType,
	},
	"account_id": schema.StringAttribute{
		Optional:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Account UUID where Workspace Role resides",
	},
	"inherited_role_id": schema.StringAttribute{
		Optional:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace Role UUID, whose permissions are inherited by this Workspace Role",
	},
}

// Schema defines the schema fro the data source.
func (d *WorkspaceRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect Workspace Role",
		Attributes:  workspaceRoleAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *WorkspaceRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *WorkspaceRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkspaceRoleDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.WorkspaceRoles(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating the Workspace Roles client",
			fmt.Sprintf("Could not create Workspace Roles client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	// Fetch an existing Workspace Role by name
	// Here, we'd expect only 1 Role (or none) to be returned
	// as we are querying a single Role name, not a list of names
	workspaceRoles, err := client.List(ctx, []string{model.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Workspace Role state",
			fmt.Sprintf("Could not read Workspace Role, unexpected error: %s", err.Error()),
		)
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

	list, diags := types.ListValueFrom(ctx, types.StringType, fetchedRole.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Permissions = list

	list, diags = types.ListValueFrom(ctx, types.StringType, fetchedRole.Scopes)
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
