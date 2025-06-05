package resources

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// BlockSchemaResource is the resource implementation.
type BlockSchemaResource struct {
	client api.PrefectClient
}

// BlockSchemaResourceModel is the model for the resource.
type BlockSchemaResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Checksum     types.String          `tfsdk:"checksum"`
	Fields       jsontypes.Normalized  `tfsdk:"fields"`
	BlockTypeID  customtypes.UUIDValue `tfsdk:"block_type_id"`
	BlockType    types.String          `tfsdk:"block_type"`
	Capabilities types.List            `tfsdk:"capabilities"`
	Version      types.String          `tfsdk:"version"`
}

// NewBlockSchemaResource returns a new BlockSchemaResource.
//
//nolint:ireturn // required by Terraform API
func NewBlockSchemaResource() resource.Resource {
	return &BlockSchemaResource{}
}

// Metadata returns the resource type name.
func (r *BlockSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_schema"
}

// Configure initializes runtime state for the resource.
func (r *BlockSchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the schema for the resource.
func (r *BlockSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `block_schema` allows creating and managing [Prefect Block Schemas](https://docs.prefect.io/latest/concepts/blocks/).",
		),
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
			"checksum": schema.StringAttribute{
				Description: "The checksum of the block schema.",
				Computed:    true,
			},
			"fields": schema.StringAttribute{
				Description: "The fields of the block schema.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"block_type_id": schema.StringAttribute{
				Description: "The ID of the block type.",
				Required:    true,
				CustomType:  customtypes.UUIDType{},
			},
			"block_type": schema.StringAttribute{
				Description: "The type of the block.",
				Computed:    true,
			},
			"capabilities": schema.ListAttribute{
				Description: "The capabilities of the block schema.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"version": schema.StringAttribute{
				Description: "The version of the block schema.",
				Optional:    true,
			},
		},
	}
}

func copyBlockSchemaToModel(ctx context.Context, blockSchema *api.BlockSchema, tfModel *BlockSchemaResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tfModel.ID = customtypes.NewUUIDValue(blockSchema.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(blockSchema.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(blockSchema.Updated)
	tfModel.Checksum = types.StringValue(blockSchema.Checksum)
	tfModel.BlockTypeID = customtypes.NewUUIDValue(blockSchema.BlockTypeID)
	tfModel.BlockType = types.StringValue(blockSchema.BlockType.Slug)
	tfModel.Version = types.StringValue(blockSchema.Version)

	// We do not persist the fields value from the API -> State
	// because the resulting value is sometimes mutated by the API, leading to
	// "inconsistent result after apply" errors.
	//
	// fields, err := json.Marshal(blockSchema.Fields)
	// if err != nil {
	// 	diags.Append(helpers.SerializeDataErrorDiagnostic("fields", "Block Schema", err))
	// }
	// tfModel.Fields = jsontypes.NewNormalizedValue(string(fields))

	capabilities, diags := types.ListValueFrom(ctx, types.StringType, blockSchema.Capabilities)
	if diags.HasError() {
		return diags
	}

	tfModel.Capabilities = capabilities

	return nil
}

// Create creates a new BlockSchema resource.
func (r *BlockSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockSchemaResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockSchemaClient, err := r.client.BlockSchemas(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Schema", err))

		return
	}

	var capabilities []string
	diags = plan.Capabilities.ElementsAs(ctx, &capabilities, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var fields interface{}
	if err := json.Unmarshal([]byte(plan.Fields.ValueString()), &fields); err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("fields", "Block Schema", err))

		return
	}

	blockSchema, err := blockSchemaClient.Create(ctx, &api.BlockSchemaCreate{
		Fields:       fields,
		BlockTypeID:  uuid.MustParse(plan.BlockTypeID.ValueString()),
		Capabilities: capabilities,
		Version:      plan.Version.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Schema", err))

		return
	}

	diags = copyBlockSchemaToModel(ctx, blockSchema, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the BlockSchema resource.
func (r *BlockSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlockSchemaResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockSchemaClient, err := r.client.BlockSchemas(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Schema", err))

		return
	}

	blockSchema, err := blockSchemaClient.Read(ctx, state.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Schema", "get", err))
	}

	diags = copyBlockSchemaToModel(ctx, blockSchema, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the BlockSchema resource.
func (r *BlockSchemaResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Update API is not implemented.
}

// Delete deletes the BlockSchema resource.
func (r *BlockSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlockSchemaResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockSchemaClient, err := r.client.BlockSchemas(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Schema", err))

		return
	}

	blockSchemaID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block Schema", err))

		return
	}

	err = blockSchemaClient.Delete(ctx, blockSchemaID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Schema", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ImportState imports the resource into Terraform state.
func (r *BlockSchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
