package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
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

	Name types.String         `tfsdk:"name"`
	Type types.String         `tfsdk:"type"`
	Data jsontypes.Normalized `tfsdk:"data"`
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
			"type": schema.StringAttribute{
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

// Create will create the Block resource through the API and insert it into the State.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read refreshes the Terraform state with the latest data.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
}

// ImportState imports the resource into Terraform state.
//
//nolint:revive // TODO: remove this comment when method is implemented
func (r *BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
