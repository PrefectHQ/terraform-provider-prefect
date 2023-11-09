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
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

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
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *WorkPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `work_pool` represents a Prefect Cloud Work Pool. " +
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
				// In general, we can use UseStateForUnknown() to avoid unnecessary
				// cases of `known after apply` states during plans. Mostly, this planmodifier
				// is suitable for Computed attributes that do not change often and
				// do not have a default value set here in the Schema.
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
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
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

// copyWorkPoolToModel copies an api.WorkPool to a WorkPoolResourceModel.
func copyWorkPoolToModel(_ context.Context, pool *api.WorkPool, model *WorkPoolResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(pool.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(pool.Created)
	model.Updated = customtypes.NewTimestampPointerValue(pool.Updated)

	model.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	model.DefaultQueueID = customtypes.NewUUIDValue(pool.DefaultQueueID)
	model.Description = types.StringPointerValue(pool.Description)
	model.Name = types.StringValue(pool.Name)
	model.Paused = types.BoolValue(pool.IsPaused)
	model.Type = types.StringValue(pool.Type)

	if pool.BaseJobTemplate != nil {
		var builder strings.Builder
		encoder := json.NewEncoder(&builder)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(pool.BaseJobTemplate)
		if err != nil {
			diags.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to serialize Base Job Template",
				fmt.Sprintf("Failed to serialize Base Job Template as JSON string: %s", err),
			)

			return diags
		}

		model.BaseJobTemplate = jsontypes.NewNormalizedValue(strings.TrimSuffix(builder.String(), "\n"))
	}

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model WorkPoolResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate := map[string]interface{}{}
	if !model.BaseJobTemplate.IsNull() {
		reader := strings.NewReader(model.BaseJobTemplate.ValueString())
		decoder := json.NewDecoder(reader)
		err := decoder.Decode(&baseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to deserialize Base Job Template",
				fmt.Sprintf("Failed to deserialize Base Job Template as JSON object: %s", err),
			)

			return
		}
	}

	client, err := r.client.WorkPools(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool client",
			fmt.Sprintf("Could not create work pool client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	pool, err := client.Create(ctx, api.WorkPoolCreate{
		Name:             model.Name.ValueString(),
		Description:      model.Description.ValueStringPointer(),
		Type:             model.Type.ValueString(),
		BaseJobTemplate:  baseJobTemplate,
		IsPaused:         model.Paused.ValueBool(),
		ConcurrencyLimit: model.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool",
			fmt.Sprintf("Could not create work pool, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(ctx, pool, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model WorkPoolResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool client",
			fmt.Sprintf("Could not create work pool client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	pool, err := client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work pool state",
			fmt.Sprintf("Could not read work pool, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(ctx, pool, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model WorkPoolResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate := map[string]interface{}{}
	if !model.BaseJobTemplate.IsNull() {
		reader := strings.NewReader(model.BaseJobTemplate.ValueString())
		decoder := json.NewDecoder(reader)
		err := decoder.Decode(&baseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to deserialize Base Job Template",
				fmt.Sprintf("Failed to deserialize Base Job Template as JSON object: %s", err),
			)

			return
		}
	}

	client, err := r.client.WorkPools(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool client",
			fmt.Sprintf("Could not create work pool client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	err = client.Update(ctx, model.Name.ValueString(), api.WorkPoolUpdate{
		Description:      model.Description.ValueStringPointer(),
		IsPaused:         model.Paused.ValueBoolPointer(),
		BaseJobTemplate:  baseJobTemplate,
		ConcurrencyLimit: model.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating work pool",
			fmt.Sprintf("Could not update work pool, unexpected error: %s", err),
		)

		return
	}

	pool, err := client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work pool state",
			fmt.Sprintf("Could not read work pool, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyWorkPoolToModel(ctx, pool, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model WorkPoolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.WorkPools(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool client",
			fmt.Sprintf("Could not create work pool client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting work pool",
			fmt.Sprintf("Could not delete work pool, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
