package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&WorkPoolResource{})
	_ = resource.ResourceWithImportState(&WorkPoolResource{})
)

// WorkPoolResource contains state for the resource.
type WorkPoolResource struct {
	client api.PrefectClient
}

// WorkPoolResourceModel defines the Terraform resource model.
type WorkPoolResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name             types.String          `tfsdk:"name"`
	Description      types.String          `tfsdk:"description"`
	Type             types.String          `tfsdk:"type"`
	Paused           types.Bool            `tfsdk:"paused"`
	ConcurrencyLimit types.Int64           `tfsdk:"concurrency_limit"`
	DefaultQueueID   customtypes.UUIDValue `tfsdk:"default_queue_id"`
	BaseJobTemplate  jsontypes.Normalized  `tfsdk:"base_job_template"`
}

// NewWorkPoolResource returns a new WorkPoolResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkPoolResource() resource.Resource {
	return &WorkPoolResource{}
}

// Metadata returns the resource type name.
func (r *WorkPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pool"
}

// Configure initializes runtime state for the resource.
func (r *WorkPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *WorkPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `work_pool` represents a Prefect Work Pool. " +
			"Work Pools represent infrastructure configurations for jobs across several common environments.\n" +
			"\n" +
			"Work Pools can be set up with default base job configurations, based on which type. " +
			"Use this in conjunction with the `prefect_worker_metadata` data source to bootstrap new Work Pools quickly.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				// We cannot use a CustomType due to a conflict with PlanModifiers; see
				// https://github.com/hashicorp/terraform-plugin-framework/issues/763
				// https://github.com/hashicorp/terraform-plugin-framework/issues/754
				Description: "Work pool ID (UUID)",
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
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider. In Prefect Cloud, either the `work_pool` resource or the provider's `workspace_id` must be set.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the work pool",
				// Work Pool names are the identifier on the API side, so
				// we do not support modifying this value. Therefore, any changes
				// to this attribute will force a replacement.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the work pool",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Default:     stringdefault.StaticString("prefect-agent"),
				Description: "Type of the work pool, eg. kubernetes, ecs, process, etc.",
				Optional:    true,
				// Work Pool types are also only set on create, and
				// we do not support modifying this value. Therefore, any changes
				// to this attribute will force a replacement.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paused": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether this work pool is paused",
				Optional:    true,
			},
			"concurrency_limit": schema.Int64Attribute{
				Description: "The concurrency limit applied to this work pool",
				Optional:    true,
			},
			"default_queue_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "The ID (UUID) of the default queue associated with this work pool",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"base_job_template": schema.StringAttribute{
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				Default:     stringdefault.StaticString("{}"),
				Description: "The base job template for the work pool, as a JSON string",
				Optional:    true,
			},
		},
	}
}

// copyWorkPoolToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
//
//nolint:ireturn // required to return Diagnostics
func copyWorkPoolToModel(pool *api.WorkPool, tfModel *WorkPoolResourceModel) diag.Diagnostic {
	tfModel.ID = types.StringValue(pool.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(pool.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(pool.Updated)

	tfModel.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	tfModel.DefaultQueueID = customtypes.NewUUIDValue(pool.DefaultQueueID)
	tfModel.Description = types.StringPointerValue(pool.Description)
	tfModel.Name = types.StringValue(pool.Name)
	tfModel.Paused = types.BoolValue(pool.IsPaused)
	tfModel.Type = types.StringValue(pool.Type)

	byteSlice, err := json.Marshal(pool.BaseJobTemplate)
	if err != nil {
		return helpers.SerializeDataErrorDiagnostic("data", "Base Job Template", err)
	}

	tfModel.BaseJobTemplate = jsontypes.NewNormalizedValue(string(byteSlice))

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkPoolResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate, diags := helpers.SafeUnmarshal(plan.BaseJobTemplate)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool", err))

		return
	}

	pool, err := client.Create(ctx, api.WorkPoolCreate{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueStringPointer(),
		Type:             plan.Type.ValueString(),
		BaseJobTemplate:  baseJobTemplate,
		IsPaused:         plan.Paused.ValueBool(),
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool", "create", err))

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkPoolResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool", err))

		return
	}

	pool, err := client.Get(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool", "get", err))

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(pool, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkPoolResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate, diags := helpers.SafeUnmarshal(plan.BaseJobTemplate)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool", err))

		return
	}

	err = client.Update(ctx, plan.Name.ValueString(), api.WorkPoolUpdate{
		Description:      plan.Description.ValueStringPointer(),
		IsPaused:         plan.Paused.ValueBoolPointer(),
		BaseJobTemplate:  baseJobTemplate,
		ConcurrencyLimit: plan.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool", "update", err))

		return
	}

	pool, err := client.Get(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool", "get", err))

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkPoolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Work Pool", err))

		return
	}

	err = client.Delete(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Work Pool", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// we'll allow input values in the form of:
	// - "workspace_id,name"
	// - "name"
	maxInputCount := 2
	inputParts := strings.Split(req.ID, ",")

	// eg. "foo,bar,baz"
	if len(inputParts) > maxInputCount {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `workspace_id,name`. Got %q", req.ID),
		)

		return
	}

	// eg. ",foo" or "foo,"
	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected non-empty import identifiers, in the form of `workspace_id,name`. Got %q", req.ID),
		)

		return
	}

	if len(inputParts) == maxInputCount {
		workspaceID := inputParts[0]
		name := inputParts[1]

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	} else {
		resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
	}
}
