package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
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

	Name     types.String         `tfsdk:"name"`
	Data     jsontypes.Normalized `tfsdk:"data"`
	TypeSlug types.String         `tfsdk:"type_slug"`
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
Get information about an existing Block by either:
- ID, or
- block type name and block name
<br>
If the ID is provided, then the block type name and block name will be ignored.
<br>
Use this data source to obtain Block-specific attributes, such as the data.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Block ID (UUID)",
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
			},
			"type_slug": schema.StringAttribute{
				Computed:    true,
				Description: "Block type slug",
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
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block", err))

		return
	}

	var block *api.BlockDocument

	switch {
	case !state.ID.IsNull():
		block, err = client.Get(ctx, state.ID.ValueUUID())
	case !state.Name.IsNull() && !state.TypeSlug.IsNull():
		block, err = client.GetByName(ctx, state.TypeSlug.ValueString(), state.Name.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Insufficient search criteria provided",
			"Provide either the ID, or the block type name and block name.",
		)

		return
	}

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block", "get", err))

		return
	}

	state.ID = customtypes.NewUUIDValue(block.ID)
	state.Created = customtypes.NewTimestampPointerValue(block.Created)
	state.Updated = customtypes.NewTimestampPointerValue(block.Updated)

	state.Name = types.StringValue(block.Name)
	state.TypeSlug = types.StringValue(block.BlockType.Slug)

	byteSlice, err := json.Marshal(block.Data)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("data", "Block Data", err))

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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}
