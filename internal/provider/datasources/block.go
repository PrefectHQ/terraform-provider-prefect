package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &blockDataSource{}
	_ datasource.DataSourceWithConfigure = &blockDataSource{}
)

// blockDataSource is the data source implementation.
type blockDataSource struct {
	client api.PrefectClient
}

// BlockDataSourceModel defines the Terraform data source model.
type BlockDataSourceModel struct {
	ID          customtypes.UUIDValue      `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name          types.String          `tfsdk:"name"`
	Data          jsontypes.Normalized  `tfsdk:"data"`
	BlockSchemaID customtypes.UUIDValue `tfsdk:"block_schema_id"`
	BlockTypeID   customtypes.UUIDValue `tfsdk:"block_type_id"`
	BlockTypeName types.String          `tfsdk:"block_type_name"`
}

// NewBlockDataSource is a helper function to simplify the provider implementation.
//
//nolint:ireturn // required by Terraform API
func NewBlockDataSource() datasource.DataSource {
	return &blockDataSource{}
}

// Metadata returns the data source type name.
func (d *blockDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block"
}

// Schema defines the scema for the data source.
func (d *blockDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Block by ID.
<br>
Use this data source to obtain Block-specific attributes, such as the data.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Block ID (UUID)",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
				Optional:    true,
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
				Optional:    true,
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the block",
				Optional:    true,
			},
			"data": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "The user-inputted Block payload, as a JSON string. The value's schema will depend on the selected `type` slug. Use `prefect block types inspect <slug>` to view the data schema for a given Block type.",
				Optional:    true,
			},
			"block_schema_id": schema.StringAttribute{
				Computed:    true,
				Description: "Block schema ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				Optional:    true,
			},
			"block_type_id": schema.StringAttribute{
				Computed:    true,
				Description: "Block type ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				Optional:    true,
			},
			"block_type_name": schema.StringAttribute{
				Computed:    true,
				Description: "Block type name",
				Optional:    true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *blockDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BlockDataSourceModel

	diag := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.BlockDocuments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating block client",
			fmt.Sprintf("Could not create block client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	var block *api.BlockDocument

	block, err = client.Get(ctx, state.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing block state",
			fmt.Sprintf("Could not read block, unexpected error: %s", err.Error()),
		)

		return
	}

	state.ID = customtypes.NewUUIDValue(block.ID)
	state.Created = customtypes.NewTimestampPointerValue(block.Created)
	state.Updated = customtypes.NewTimestampPointerValue(block.Updated)

	state.Name = types.StringValue(block.Name)
	state.BlockSchemaID = customtypes.NewUUIDValue(block.BlockSchemaID)
	state.BlockTypeID = customtypes.NewUUIDValue(block.BlockTypeID)
	state.BlockTypeName = types.StringPointerValue(block.BlockTypeName)

	byteSlice, err := json.Marshal(block.Data)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("data"),
			"Failed to serialize Block Data",
			fmt.Sprintf("Could not serialize Block Data as JSON string: %s", err.Error()),
		)

		return
	}

	state.Data = jsontypes.NewNormalizedValue(string(byteSlice))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure initializes runtime state for the data source.
func (d *blockDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
