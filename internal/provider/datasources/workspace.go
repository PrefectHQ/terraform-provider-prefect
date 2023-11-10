package datasources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		Description: "Workspace ID (UUID)",
		Computed:    true,
		Optional:    true,
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
	"account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID), defaults to the account set in the provider",
		Optional:    true,
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name of the workspace",
	},
	"handle": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
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
		Description: `
Get information about an existing Workspace by handle.
<br>
Use this data source to obtain Workspace IDs
`,
		Attributes: workspaceAttributes,
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

	if model.ID.IsNull() && model.Handle.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Handle are unset",
			"Either a Workspace ID or Handle are required to read a workspace.",
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

	// A workspace can be read by either ID or Handle
	// If both are set, we prefer the ID
	// If neither are set, we will fail early.
	var workspace *api.Workspace
	if !model.ID.IsNull() {
		var workspaceID uuid.UUID
		workspaceID, err = uuid.Parse(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error parsing Workspace ID",
				fmt.Sprintf("Could not parse workspace ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}

		workspace, err = client.Get(ctx, workspaceID)
	} else if !model.Handle.IsNull() {
		var workspaces []*api.Workspace
		workspaces, err = client.List(ctx, []string{model.Handle.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific workspace is not returned
		if err == nil && len(workspaces) != 1 {
			err = fmt.Errorf("a workspace with the handle=%s could not be found", model.Handle.ValueString())
		}

		if len(workspaces) == 1 {
			workspace = workspaces[0]
		}
	}

	if workspace == nil {
		resp.Diagnostics.AddError(
			"Error refreshing workspace state",
			fmt.Sprintf("Could not find workspace with ID=%s and Handle=%s", model.ID.ValueString(), model.Handle.ValueString()),
		)

		return
	}

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
