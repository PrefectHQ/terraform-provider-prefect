package resources

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// BlockTypeResource is the resource implementation for block types.
type BlockTypeResource struct {
	client api.PrefectClient
}

// BlockTypeResourceModel is the model for the resource.
type BlockTypeResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name             types.String `tfsdk:"name"`
	Slug             types.String `tfsdk:"slug"`
	LogoURL          types.String `tfsdk:"logo_url"`
	DocumentationURL types.String `tfsdk:"documentation_url"`
	Description      types.String `tfsdk:"description"`
	CodeExample      types.String `tfsdk:"code_example"`

	IsProtected types.Bool `tfsdk:"is_protected"`
}

// NewBlockTypeResource returns a new BlockTypeResource.
//
//nolint:ireturn // required by Terraform API
func NewBlockTypeResource() resource.Resource {
	return &BlockTypeResource{}
}

// Metadata returns the resource type name.
func (r *BlockTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_type"
}

// Configure initializes runtime state for the resource.
func (r *BlockTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *BlockTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `block_type` allows creating and managing [Prefect Block Types](https://docs.prefect.io/latest/concepts/blocks/).",
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Block Type ID (UUID)",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the block type.",
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "The slug of the block type.",
			},
			"logo_url": schema.StringAttribute{
				Optional:    true,
				Description: "Web URL for the block type's logo.",
			},
			"documentation_url": schema.StringAttribute{
				Optional:    true,
				Description: "Web URL for the block type's documentation.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A short blurb about the corresponding block's intended use.",
			},
			"code_example": schema.StringAttribute{
				Optional:    true,
				Description: "A code snippet demonstrating use of the corresponding block.",
			},
			"is_protected": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the block type is protected. Protected block types cannot be modified via API.",
			},
		},
	}
}

// copyBlockTypeToModel copies the block type to the model.
func copyBlockTypeToModel(blockType *api.BlockType, tfModel *BlockTypeResourceModel) {
	tfModel.ID = customtypes.NewUUIDValue(blockType.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(blockType.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(blockType.Updated)

	tfModel.Name = types.StringValue(blockType.Name)
	tfModel.Slug = types.StringValue(blockType.Slug)

	tfModel.LogoURL = types.StringValue(blockType.LogoURL)
	tfModel.DocumentationURL = types.StringValue(blockType.DocumentationURL)
	tfModel.Description = types.StringValue(blockType.Description)
	tfModel.CodeExample = types.StringValue(blockType.CodeExample)
	tfModel.IsProtected = types.BoolValue(blockType.IsProtected)
}

// Create creates a new block type.
func (r *BlockTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockTypeResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeClient, err := r.client.BlockTypes(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Type", err))

		return
	}

	blockType, err := blockTypeClient.Create(ctx, &api.BlockTypeCreate{
		Name:             plan.Name.ValueString(),
		Slug:             plan.Slug.ValueString(),
		LogoURL:          plan.LogoURL.ValueString(),
		DocumentationURL: plan.DocumentationURL.ValueString(),
		Description:      plan.Description.ValueString(),
		CodeExample:      plan.CodeExample.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Type", err))

		return
	}

	copyBlockTypeToModel(blockType, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the BlockType resource.
func (r *BlockTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlockTypeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeClient, err := r.client.BlockTypes(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Type", err))

		return
	}

	blockType, err := blockTypeClient.GetBySlug(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Type", "get by slug", err))

		return
	}

	copyBlockTypeToModel(blockType, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the BlockType resource.
func (r *BlockTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BlockTypeResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeClient, err := r.client.BlockTypes(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Type", err))

		return
	}

	var state BlockTypeResourceModel

	// Get the current state so we can use the ID to update the block type.
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Block Type", err))

		return
	}

	err = blockTypeClient.Update(ctx, blockTypeID, &api.BlockTypeUpdate{
		LogoURL:          plan.LogoURL.ValueString(),
		DocumentationURL: plan.DocumentationURL.ValueString(),
		Description:      plan.Description.ValueString(),
		CodeExample:      plan.CodeExample.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Type", "update", err))

		return
	}

	blockType, err := blockTypeClient.GetBySlug(ctx, plan.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Type", "get by slug", err))

		return
	}

	copyBlockTypeToModel(blockType, &plan)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the BlockType resource.
func (r *BlockTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlockTypeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	blockTypeClient, err := r.client.BlockTypes(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Block Type", err))

		return
	}

	err = blockTypeClient.Delete(ctx, state.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Block Type", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ImportState imports the resource into Terraform state.
func (r *BlockTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
