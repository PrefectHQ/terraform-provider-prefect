package resources

import (
	"context"

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

var (
	_ = resource.ResourceWithConfigure(&ConcurrencyLimitResource{})
	_ = resource.ResourceWithImportState(&ConcurrencyLimitResource{})
)

// ConcurrencyLimitResource contains state for the resource.
type ConcurrencyLimitResource struct {
	client api.PrefectClient
}

// ConcurrencyLimitResourceModel defines the Terraform resource model.
type ConcurrencyLimitResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Tag              types.String `tfsdk:"tag"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
}

// NewConcurrencyLimitResource returns a new ConcurrencyLimitResource.
//
//nolint:revive // we use the resource.ResourceWithConfigure helper instead
func NewConcurrencyLimitResource() resource.Resource {
	return &ConcurrencyLimitResource{}
}

// Metadata returns the resource type name.
func (r *ConcurrencyLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_concurrency_limit"
}

// Configure initializes runtime state for the resource.
func (r *ConcurrencyLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema defines the schema for the resource.
func (r *ConcurrencyLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `concurrency_limit` represents a concurrency limit for a tag.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Concurrency limit ID (UUID)",
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
				Required:    true,
				Description: "Account ID (UUID)",
			},
			"workspace_id": schema.StringAttribute{
				Required:    true,
				Description: "Workspace ID (UUID)",
			},
			"tag": schema.StringAttribute{
				Required:    true,
				Description: "Tag",
			},
			"concurrency_limit": schema.Int64Attribute{
				Required:    true,
				Description: "Concurrency limit",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ConcurrencyLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConcurrencyLimitResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ConcurrencyLimits(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Concurrency Limit", err))

		return
	}

	concurrencyLimit, err := client.Create(ctx, api.ConcurrencyLimitCreate{
		Tag:              plan.Tag.ValueString(),
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Concurrency Limit", "create", err))

		return
	}

	copyConcurrencyLimitToModel(concurrencyLimit, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func copyConcurrencyLimitToModel(concurrencyLimit *api.ConcurrencyLimit, model *ConcurrencyLimitResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(concurrencyLimit.ID.String())
	model.Created = customtypes.NewTimestampValue(*concurrencyLimit.Created)
	model.Updated = customtypes.NewTimestampValue(*concurrencyLimit.Updated)
	model.Tag = types.StringValue(concurrencyLimit.Tag)
	model.ConcurrencyLimit = types.Int64Value(concurrencyLimit.ConcurrencyLimit)

	return nil
}

// Delete deletes the resource.
func (r *ConcurrencyLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ConcurrencyLimits(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Concurrency Limit", err))

		return
	}

	err = client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Concurrency Limit", "delete", err))

		return
	}
}
