package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

type BlockResource struct {
	client api.PrefectClient
}

type BlockResourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name     types.String         `tfsdk:"name"`
	TypeSlug types.String         `tfsdk:"type_slug"`
	Data     jsontypes.Normalized `tfsdk:"data"`
}

// NewBlockResource returns a new BlockResource.
//
//nolint:ireturn // required by Terraform API
func NewBlockResource() resource.Resource {
	return &BlockResource{}
}

// Metadata returns the resource type name.
func (r *BlockResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block"
}

// Configure initializes runtime state for the resource.
func (r *BlockResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *BlockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `block` allows creating and managing [Prefect Blocks](https://docs.prefect.io/latest/concepts/blocks/), " +
			"which are primitives for configuration / secrets in your flows." +
			"\n" +
			"`block` resources represent configurations for schemas for all different Block types. " +
			"Because of the polymorphic nature of Blocks, you should utilize the `prefect` [CLI](https://docs.prefect.io/latest/getting-started/installation/) to inspect all Block types and schemas." +
			"\n" +
			"Use `prefect block types ls` to view all available Block type slugs, which is used in the `type` attribute." +
			"\n" +
			"Use `prefect block types inspect <slug>` to view the data schema for a given Block type. Use this to construct the `data` attribute value (as JSON string).",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Block ID (UUID)",
				// attributes which are not configurable + should not show updates from the existing state value
				// should implement `UseStateForUnknown()`
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Description: "Unique name of the Block",
			},
			"type_slug": schema.StringAttribute{
				Required:    true,
				Description: "Block Type slug, which determines the schema of the `data` JSON attribute. Use `prefect block types ls` to view all available Block type slugs.",
			},
			"data": schema.StringAttribute{
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "The user-inputted Block payload, as a JSON string. The value's schema will depend on the selected `type` slug. Use `prefect block types inspect <slug>` to view the data schema for a given Block type.",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the Block is located",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) where the Block is located",
			},
		},
	}
}

// copyBlockToModel maps an API response to a model that is saved in Terraform state.
func copyBlockToModel(_ context.Context, block *api.BlockDocument, state *BlockResourceModel) diag.Diagnostics {
	state.ID = types.StringValue(block.ID.String())
	state.Created = customtypes.NewTimestampPointerValue(block.Created)
	state.Updated = customtypes.NewTimestampPointerValue(block.Updated)
	state.Name = types.StringValue(block.Name)
	state.TypeSlug = types.StringValue(block.BlockType.Slug)
	state.Data = jsontypes.NewNormalizedValue(block.Data)

	// NOTE:
	// we can perform deserializing checks here
	// and optionally return an error diagnostic.
	return nil
}

// Create will create the Block resource through the API and insert it into the State.
func (r *BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config BlockResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeClient, err := r.client.BlockTypes(config.AccountID.ValueUUID(), config.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Types", err))
	}
	blockSchemaClient, err := r.client.BlockSchemas(config.AccountID.ValueUUID(), config.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Schema", err))
	}
	blockDocumentClient, err := r.client.BlockDocuments(config.AccountID.ValueUUID(), config.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))
	}

	// fetch the BlockType by the provided `type_slug`
	blockType, err := blockTypeClient.GetBySlug(ctx, config.TypeSlug.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Type", "get_by_slug", err))

		return
	}

	blockSchemas, err := blockSchemaClient.List(ctx, []uuid.UUID{blockType.ID})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Schema", "list", err))

		return
	}

	if len(blockSchemas) == 0 {
		resp.Diagnostics.AddError(
			"No block schemas found",
			"No block schemas found for the provided type slug",
		)

		return
	}

	latestBlockSchema := blockSchemas[0]

	createdBlockDocument, err := blockDocumentClient.Create(ctx, api.BlockDocumentCreate{
		Name:          config.Name.ValueString(),
		Data:          config.Data.ValueString(),
		BlockSchemaID: latestBlockSchema.ID,
		BlockTypeID:   latestBlockSchema.BlockTypeID,
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "create", err))

		return
	}

	resp.Diagnostics.Append(copyBlockToModel(ctx, createdBlockDocument, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *BlockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlockResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		resp.Diagnostics.AddError(
			"ID is unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)
	}

	client, err := r.client.BlockDocuments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating block client",
			fmt.Sprintf("Could not create block client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	var blockID uuid.UUID
	blockID, err = uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Block ID",
			fmt.Sprintf("Could not parse block ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	block, err := client.Get(ctx, blockID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing block state",
			fmt.Sprintf("Could not read block, unexpected error: %s", err.Error()),
		)

		return
	}

	copyBlockToModel(ctx, block, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlockResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockDocumentClient, err := r.client.BlockDocuments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	blockDocumentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Block ID",
			fmt.Sprintf("Could not parse block ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = blockDocumentClient.Delete(ctx, blockDocumentID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
