package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

type BlockResource struct {
	client api.PrefectClient
}

type BlockResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name          types.String         `tfsdk:"name"`
	TypeSlug      types.String         `tfsdk:"type_slug"`
	Data          jsontypes.Normalized `tfsdk:"data"`
	DataWO        jsontypes.Normalized `tfsdk:"data_wo"`
	DataWOVersion types.Int32          `tfsdk:"data_wo_version"`
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
		Description: helpers.DescriptionWithPlans(
			"The resource `block` allows creating and managing [Prefect Blocks](https://docs.prefect.io/latest/concepts/blocks/), "+
				"which are primitives for configuration / secrets in your flows."+
				"\n"+
				"`block` resources represent configurations for schemas for all different Block types. "+
				"Because of the polymorphic nature of Blocks, you should utilize the `prefect` [CLI](https://docs.prefect.io/latest/getting-started/installation/) to inspect all Block types and schemas."+
				"\n"+
				"*Note:* you should be on version `3.0.0rc1` or later to use the following commands:"+
				"\n"+
				"Use `prefect block type ls` to view all available Block type slugs, which is used in the `type_slug` attribute."+
				"\n"+
				"Use `prefect block type inspect <slug>` to view the data schema for a given Block type. Use this to construct the `data` attribute value (as JSON string)."+
				"\n"+
				"*NOTE:* drift in the `.data` attribute is detected for Blocks managed in Terraform, so out-of-band changes (e.g. edits made in the Prefect UI) will surface on the next plan. Two carve-outs apply: (1) Blocks whose `data` contains a `$ref` expression -- the syntax used to reference another Block document by ID, e.g. an `s3-bucket` Block pointing at an `aws-credentials` Block, or a `dbt-core-operation` pointing at a `dbt-cli-profile` (see the `$ref` example below) -- do not have drift detected, because the API resolves `$ref` server-side and returns the resolved nested data, which would never match the literal HCL. (2) Blocks managed via the write-only `data_wo` attribute also do not have drift detected, because state must remain null to honor the write-only contract; bump `data_wo_version` to force an update in that case.",
			helpers.AllPlans...,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type_slug": schema.StringAttribute{
				Required:    true,
				Description: "Block Type slug, which determines the schema of the `data` JSON attribute. Use `prefect block type ls` to view all available Block type slugs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "The user-inputted Block payload, as a JSON string. Use `jsonencode` on the provided value to satisfy the underlying JSON type. The value's schema will depend on the selected `type` slug. Use `prefect block type inspect <slug>` to view the data schema for a given Block type.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("data_wo"),
					),
				},
			},
			"data_wo": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "The user-inputted Block payload, as a JSON string. Use `jsonencode` on the provided value to satisfy the underlying JSON type. The value's schema will depend on the selected `type` slug. Use `prefect block type inspect <slug>` to view the data schema for a given Block type.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("data"),
					),
				},
			},
			"data_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the `data_wo` attribute. This is used to track changes to the `data_wo` attribute and trigger updates when the value changes.",
				Validators: []validator.Int32{
					int32validator.AlsoRequires(path.MatchRoot("data_wo")),
				},
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID) where the Block is located",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID) where the Block is located. In Prefect Cloud, either the `prefect_block` resource or the provider's `workspace_id` must be set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// getBlockSchemas fetches the block schemas for a given block type slug.
//
//nolint:ireturn // required by Terraform API
func (r *BlockResource) getBlockSchemas(ctx context.Context, plan BlockResourceModel) ([]*api.BlockSchema, diag.Diagnostic) {
	blockTypeClient, err := r.client.BlockTypes(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		return nil, helpers.CreateClientErrorDiagnostic("Block Types", err)
	}

	blockSchemaClient, err := r.client.BlockSchemas(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		return nil, helpers.CreateClientErrorDiagnostic("Block Schema", err)
	}

	blockType, err := blockTypeClient.GetBySlug(ctx, plan.TypeSlug.ValueString())
	if err != nil {
		return nil, helpers.ResourceClientErrorDiagnostic("Block Type", "get_by_slug", err)
	}

	blockSchemas, err := blockSchemaClient.List(ctx, []uuid.UUID{blockType.ID})
	if err != nil {
		return nil, helpers.ResourceClientErrorDiagnostic("Block Schema", "list", err)
	}

	return blockSchemas, nil
}

// getLatestBlockSchema fetches the latest block schema for a given block type slug.
// If no block schemas are returned, the retrieval is retried because Prefect creates
// them asynchronously after the creation of a workspace.
//
//nolint:ireturn // required by Terraform API
func (r *BlockResource) getLatestBlockSchema(ctx context.Context, plan BlockResourceModel) (*api.BlockSchema, diag.Diagnostic) {
	var blockSchemas []*api.BlockSchema
	var latestBlockSchema *api.BlockSchema
	var diags diag.Diagnostic

	err := retry.Do(func() error {
		blockSchemas, diags = r.getBlockSchemas(ctx, plan)
		if diags != nil {
			return fmt.Errorf("unable to get block schemas: %s", diags.Detail())
		}

		if len(blockSchemas) == 0 {
			return fmt.Errorf("no block schemas found")
		}

		latestBlockSchema = blockSchemas[0]

		return nil
	})

	if err != nil {
		diags = diag.NewErrorDiagnostic(
			"No block schemas found",
			fmt.Sprintf("No block schemas found for %s block type slug", plan.TypeSlug.ValueString()),
		)

		return nil, diags
	}

	return latestBlockSchema, nil
}

// copyBlockToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
//
// When persistData is true, the API's Data payload is serialized and copied onto
// tfModel.Data so that out-of-band changes can be detected as drift. Callers must
// pass false when the user supplied a $ref expression (the API resolves $ref
// server-side and would return the resolved form, causing inconsistent-result
// errors) or when the user supplied data via the write-only data_wo attribute
// (state must remain null to honor the write-only contract).
func copyBlockToModel(block *api.BlockDocument, tfModel *BlockResourceModel, persistData bool) diag.Diagnostics {
	var diags diag.Diagnostics

	tfModel.ID = customtypes.NewUUIDValue(block.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(block.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(block.Updated)
	tfModel.Name = types.StringValue(block.Name)
	tfModel.TypeSlug = types.StringValue(block.BlockType.Slug)

	if persistData {
		byteSlice, err := json.Marshal(block.Data)
		if err != nil {
			diags.Append(helpers.SerializeDataErrorDiagnostic("data", "Block Data", err))

			return diags
		}
		tfModel.Data = jsontypes.NewNormalizedValue(string(byteSlice))
	}

	return diags
}

// containsRef reports whether the decoded block data contains a "$ref" key
// at any depth. Block payloads can reference other blocks via {"$ref": ...};
// the API resolves these server-side, so a GET returns the resolved form
// rather than the literal $ref expression. We must avoid persisting the
// resolved form back to state, otherwise Terraform reports an inconsistent
// result after apply and a permanent diff against the user's HCL.
func containsRef(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		if _, ok := v["$ref"]; ok {
			return true
		}
		for _, child := range v {
			if containsRef(child) {
				return true
			}
		}
	case []any:
		for _, child := range v {
			if containsRef(child) {
				return true
			}
		}
	}

	return false
}

// Create will create the Block resource through the API and insert it into the State.
func (r *BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Also get the config to evaluate write-only attributes that
	// are only available in the config, not the plan.
	var config BlockResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockDocumentClient, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	latestBlockSchema, blockSchemaDiags := r.getLatestBlockSchema(ctx, plan)
	resp.Diagnostics.Append(blockSchemaDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	// We typed `data` as JSON, as this is the most
	// flexible way to handle a dynamic schema from the API.
	// Here, we unmarshal the user-provided `data` JSON string to a map[string]interface{}
	// because we'll later need to re-marshall the entire BlockDocumentCreate payload
	// when sending it back up to the API
	data, diags := helpers.UnmarshalOptional(plan.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataWO, diags := helpers.UnmarshalOptional(config.DataWO)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(dataWO) != 0 {
		data = dataWO
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

	// Persist the API's data payload to state so that out-of-band changes show
	// up as drift on subsequent plans. Skip when:
	//   - data_wo was used (write-only contract requires state to stay null), or
	//   - the user supplied a $ref (API resolves it server-side; persisting the
	//     resolved form would cause inconsistent-result errors), or
	//   - the user supplied no data at all.
	persistData := len(dataWO) == 0 && len(data) > 0 && !containsRef(data)

	// The POST response does not include unmasked secrets, so fetch the block
	// again before copying its Data into state.
	if persistData {
		blockID := createdBlockDocument.ID
		fetched, err := blockDocumentClient.Get(ctx, blockID)
		if err != nil {
			resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block", "get", err))

			return
		}
		createdBlockDocument = fetched
	}

	diags = copyBlockToModel(createdBlockDocument, &plan, persistData)
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

		return
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
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block", "get", err))

		return
	}

	// Decide whether to persist Data based on the prior state: skip when the
	// user used data_wo (state.Data is null) or when the prior state contained
	// a $ref expression that the API would have resolved server-side.
	priorData, priorDataDiags := helpers.UnmarshalOptional(state.Data)
	resp.Diagnostics.Append(priorDataDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	persistData := !state.Data.IsNull() && !containsRef(priorData)

	diags = copyBlockToModel(block, &state, persistData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// Also get the config to evaluate write-only attributes that
	// are only available in the config, not the plan.
	var config BlockResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Also retrieve the state to compare the data_wo_version.
	var state BlockResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockDocumentClient, err := r.client.BlockDocuments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Document", err))

		return
	}

	latestBlockSchema, blockSchemaDiags := r.getLatestBlockSchema(ctx, plan)
	resp.Diagnostics.Append(blockSchemaDiags)
	if resp.Diagnostics.HasError() {
		return
	}

	blockID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block", err))

		return
	}

	data, diags := helpers.UnmarshalOptional(plan.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the data_wo_version is different, we need to update the data payload
	// using the content of the data_wo attribute.
	if !plan.DataWOVersion.Equal(state.DataWOVersion) {
		data, diags = helpers.UnmarshalOptional(config.DataWO)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
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

	// Persist the API's data payload to state when safe to do so. See the
	// matching block in Create for the full rationale.
	persistData := plan.DataWOVersion.Equal(state.DataWOVersion) &&
		!plan.Data.IsNull() &&
		!containsRef(data)

	diags = copyBlockToModel(block, &plan, persistData)
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
func (r *BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
