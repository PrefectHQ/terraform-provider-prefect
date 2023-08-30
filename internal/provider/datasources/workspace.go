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

var _ = datasource.DataSourceWithConfigure(&WorkspaceDataSource{})

// WorkspaceDataSource contains state for the data source.
type WorkspaceDataSource struct {
	client api.PrefectClient
}

// WorkspaceDataSourceModel defines the Terraform data source model.
type WorkspaceDataSourceModel struct {
	ID        customtypes.UUIDValue      `tfsdk:"id"`
	Created   customtypes.TimestampValue `tfsdk:"created"`
	Updated   customtypes.TimestampValue `tfsdk:"updated"`
	AccountID customtypes.UUIDValue      `tfsdk:"account_id"`

	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
}

// NewWorkspaceDataSource returns a new WorkspaceDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceDataSource() datasource.DataSource {
	return &WorkspaceDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

// Configure initializes runtime state for the data source.
func (d *WorkspaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

var workspaceAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace UUID",
		Required:    true,
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the workspace creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the workspace was last updated in RFC 3339 format",
	},
	"account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account UUID, defaults to the account set in the provider",
		Optional:    true,
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name of the workspace",
	},
	"handle": schema.StringAttribute{
		Computed:    true,
		Description: "Unique handle for the workspace",
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description for the workspace",
	},
}

// Schema defines the schema for the data source.
func (d *WorkspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect workspace",
		Attributes:  workspaceAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkspaceDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !model.ID.IsNull() && !model.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting workspace lookup keys",
			"Workspaces can be identified by their unique name or ID, but not both.",
		)

		return
	}

	client, err := d.client.Workspaces(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	workspace, err := client.Get(ctx, model.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing workspace state",
			fmt.Sprintf("Could not read workspace, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(workspace.ID)
	model.Created = customtypes.NewTimestampPointerValue(workspace.Created)
	model.Updated = customtypes.NewTimestampPointerValue(workspace.Updated)

	model.Name = types.StringValue(workspace.Name)
	model.Handle = types.StringValue(workspace.Handle)
	model.Description = types.StringPointerValue(workspace.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
