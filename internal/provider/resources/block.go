package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

const (
	blockSchemaRetryCount = 3
	blockSchemaRetryDelay = 2 * time.Second
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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

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
			"*Note:* you should be on version `3.0.0rc1` or later to use the following commands:" +
			"\n" +
			"Use `prefect block type ls` to view all available Block type slugs, which is used in the `type_slug` attribute." +
			"\n" +
			"Use `prefect block type inspect <slug>` to view the data schema for a given Block type. Use this to construct the `data` attribute value (as JSON string).",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Block ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Description: "Block Type slug, which determines the schema of the `data` JSON attribute. Use `prefect block type ls` to view all available Block type slugs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "The user-inputted Block payload, as a JSON string. Use `jsonencode` on the provided value to satisfy the underlying JSON type. The value's schema will depend on the selected `type` slug. Use `prefect block type inspect <slug>` to view the data schema for a given Block type.",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the Block is located",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) where the Block is located. In Prefect Cloud, either the `prefect_block` resource or the provider's `workspace_id` must be set.",
			},
		},
	}
}

// copyBlockToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyBlockToModel(block *api.BlockDocument, tfModel *BlockResourceModel) diag.Diagnostics {
	// NOTE: we will map the `data` key OUTSIDE of this helper function, as we will
	// need to skip this step for the POST /block_documents endpoint,
	// which always returns masked data + will create inconsistent state between the
	// plan <> fetched value - for Create(), we'll fall back to the user-configured JSON payload,
	// whereas for Read() / Update() we can ask the API for unmasked values to ensure a consistent
	// state drift check.
	tfModel.ID = types.StringValue(block.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(block.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(block.Updated)
	tfModel.Name = types.StringValue(block.Name)
	tfModel.TypeSlug = types.StringValue(block.BlockType.Slug)

	return nil
}

// getLatestBlockSchema retrieves the latest block schema for a given block type slug.
//
// BlockTypes / BlockSchemas are tied to a particular Workspace, and
// they are created asynchronously after a new Workspace request resolves.
//
// This means that if a `prefect_block` is created in the same plan as a
// new `prefect_workspace`, there is a potential race condition that could occur if the
// Workspace is created + resolves, but the BlockTypes / BlockSchemas have yet to have
// been created - this has been observed to happen even with a depends_on relationship defined.
//
// To eliminate spurious create failures, test flakiness, and general user confusion,
// we'll add a retry for these specific operations.
//
//nolint:ireturn // returns diagnostic for simpler error handling
func getLatestBlockSchema(ctx context.Context, client api.PrefectClient, plan BlockResourceModel) (api.BlockSchema, diag.Diagnostic) {
	var result api.BlockSchema

	blockTypeClient, err := client.BlockTypes(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		return api.BlockSchema{}, helpers.CreateClientErrorDiagnostic("Block Types", err)
	}

	blockSchemaClient, err := client.BlockSchemas(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		return api.BlockSchema{}, helpers.CreateClientErrorDiagnostic("Block Schema", err)
	}

	blockType, err := blockTypeClient.GetBySlug(ctx, plan.TypeSlug.ValueString())
	if err != nil {
		return api.BlockSchema{}, helpers.ResourceClientErrorDiagnostic("Block Type", "get_by_slug", err)
	}

	listBlockSchemasFunc := func() ([]*api.BlockSchema, diag.Diagnostic) {
		blockSchemas, err := blockSchemaClient.List(ctx, []uuid.UUID{blockType.ID})
		if err != nil {
			return []*api.BlockSchema{}, helpers.ResourceClientErrorDiagnostic("Block Schema", "list", err)
		}

		return blockSchemas, nil
	}

	for i := 0; i < blockSchemaRetryCount; i++ {
		// Retrieve the block schemas.
		blockSchemas, err := listBlockSchemasFunc()
		if err != nil {
			return api.BlockSchema{}, err
		}

		// If block schemas are found, set the latest block schema and break out of the loop.
		if len(blockSchemas) > 0 {
			result = *blockSchemas[0]

			break
		}

		// Sleep for a short duration to allow the API to catch up.
		time.Sleep(blockSchemaRetryDelay)
	}

	// If the latest block schema is still nil, set the error diagnostic.
	if result.ID == uuid.Nil {
		return api.BlockSchema{}, diag.NewErrorDiagnostic(
			"No block schemas found",
			fmt.Sprintf("No block schemas found for %s block type slug", plan.TypeSlug.ValueString()),
		)
	}

	return result, nil
}

// Create will create the Block resource through the API and insert it into the State.
func (r *BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	latestBlockSchema, blockSchemaDiags := getLatestBlockSchema(ctx, r.client, plan)
	resp.Diagnostics.Append(blockSchemaDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	// We typed `data` as JSON, as this is the most
	// flexible way to handle a dynamic schema from the API.
	// Here, we unmarshal the user-provided `data` JSON string to a map[string]interface{}
	// because we'll later need to re-marshall the entire BlockDocumentCreate payload
	// when sending it back up to the API
	var data map[string]interface{}
	resp.Diagnostics.Append(plan.Data.Unmarshal(&data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockDocumentClient, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	createdBlockDocument, err := blockDocumentClient.Create(ctx, api.BlockDocumentCreate{
		Name:          plan.Name.ValueString(),
		Data:          data,
		BlockSchemaID: latestBlockSchema.ID,
		BlockTypeID:   latestBlockSchema.BlockTypeID,
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "create", err))

		return
	}

	diags = copyBlockToModel(createdBlockDocument, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	var blockID uuid.UUID
	blockID, err = uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block", err))

		return
	}

	block, err := client.Get(ctx, blockID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block", "get", err))

		return
	}

	diags = copyBlockToModel(block, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	byteSlice, err := json.Marshal(block.Data)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("data", "Block Data", err))

		return
	}
	state.Data = jsontypes.NewNormalizedValue(string(byteSlice))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BlockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BlockResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	latestBlockSchema, blockSchemaDiags := getLatestBlockSchema(ctx, r.client, plan)
	resp.Diagnostics.Append(blockSchemaDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	blockID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block", err))

		return
	}

	var data map[string]interface{}
	resp.Diagnostics.Append(plan.Data.Unmarshal(&data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockDocumentClient, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	err = blockDocumentClient.Update(ctx, blockID, api.BlockDocumentUpdate{
		BlockSchemaID: latestBlockSchema.ID,
		Data:          data,

		// NOTE: setting this to `false` will replace the contents of `.data`
		// We want to do this on Update() - if we don't, removing top-level keys
		// will cause the API to ignore those removals, which causes a provider-level
		// state conflict + failure.
		MergeExistingData: false,
	})

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Document", "update", err))

		return
	}

	block, err := blockDocumentClient.Get(ctx, blockID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block", "get", err))

		return
	}

	diags := copyBlockToModel(block, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normally, we would also copy the retrieved Block's Data field into the
	// plan object before setting the current state.
	//
	// However, the API's GET method does not return the `$ref` expression if it
	// was specified in the Data field on the Block resource. This leads to
	// "inconsistent result after apply" errors. For now, we'll skip copying the
	// retrieved Block's Data field and use what was specified in the plan.

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
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
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block", err))

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
// Valid import IDs:
// <block_id>
// <block_id>,<workspace_id>.
func (r *BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) > 2 || len(parts) == 0 {
		resp.Diagnostics.AddError(
			"Error importing Block",
			"Import ID must be in the format of <block identifier> OR <block identifier>,<workspace_id>",
		)

		return
	}

	blockIdentifier := parts[0]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), blockIdentifier)...)

	if len(parts) == 2 && parts[1] != "" {
		workspaceID, err := uuid.Parse(parts[1])
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID.String())...)
	}
}
