package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&GlobalConcurrencyLimitResource{})
	_ = resource.ResourceWithImportState(&GlobalConcurrencyLimitResource{})
)

// GlobalConcurrencyLimitResource contains state for the resource.
type GlobalConcurrencyLimitResource struct {
	client api.PrefectClient
}

// GlobalConcurrencyLimitResourceModel defines the Terraform resource model.
type GlobalConcurrencyLimitResourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name   types.String `tfsdk:"name"`
	Limit  types.Int64  `tfsdk:"limit"`
	Active types.Bool   `tfsdk:"active"`

	ActiveSlots        types.Int64   `tfsdk:"active_slots"`
	SlotDecayPerSecond types.Float64 `tfsdk:"slot_decay_per_second"`
}

// NewGlobalConcurrencyLimitResource returns a new GlobalConcurrencyLimitResource.
//
//nolint:ireturn // required by Terraform API
func NewGlobalConcurrencyLimitResource() resource.Resource {
	return &GlobalConcurrencyLimitResource{}
}

// Metadata returns the resource type name.
func (r *GlobalConcurrencyLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_concurrency_limit"
}

// Configure initializes runtime state for the resource.
func (r *GlobalConcurrencyLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *GlobalConcurrencyLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `global_concurrency_limit` represents a global concurrency limit. Global concurrency limits allow you to control how many tasks can run simultaneously across all workspaces. For more information, see [apply global concurrency and rate limits](https://docs.prefect.io/v3/develop/global-concurrency-limits).",
			helpers.AllPlans...,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Global concurrency limit ID (UUID)",
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
				Description: "Account ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "Workspace ID (UUID)",
				CustomType:  customtypes.UUIDType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the global concurrency limit.",
			},
			"limit": schema.Int64Attribute{
				Required:    true,
				Description: "The maximum number of tasks that can run simultaneously.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the global concurrency limit is active.",
				Default:     booldefault.StaticBool(true),
			},
			"active_slots": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The number of active slots.",
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"slot_decay_per_second": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Slot Decay Per Second (number or null)",
				Default:     float64default.StaticFloat64(0),
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
		},
	}
}

// Create creates a new global concurrency limit.
func (r *GlobalConcurrencyLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GlobalConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GlobalConcurrencyLimits(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Global Concurrency Limit", err))

		return
	}

	globalConcurrencyLimit, err := client.Create(ctx, api.GlobalConcurrencyLimitCreate{
		Name:               plan.Name.ValueString(),
		Limit:              plan.Limit.ValueInt64(),
		Active:             plan.Active.ValueBool(),
		ActiveSlots:        plan.ActiveSlots.ValueInt64(),
		SlotDecayPerSecond: plan.SlotDecayPerSecond.ValueFloat64(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Global Concurrency Limit", "create", err))

		return
	}

	copyGlobalConcurrencyLimitToModel(globalConcurrencyLimit, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func copyGlobalConcurrencyLimitToModel(globalConcurrencyLimit *api.GlobalConcurrencyLimit, model *GlobalConcurrencyLimitResourceModel) diag.Diagnostics {
	model.ID = customtypes.NewUUIDValue(globalConcurrencyLimit.ID)
	model.Created = customtypes.NewTimestampValue(*globalConcurrencyLimit.Created)
	model.Updated = customtypes.NewTimestampValue(*globalConcurrencyLimit.Updated)
	model.Name = types.StringValue(globalConcurrencyLimit.Name)
	model.Limit = types.Int64Value(globalConcurrencyLimit.Limit)
	model.Active = types.BoolValue(globalConcurrencyLimit.Active)
	model.ActiveSlots = types.Int64Value(globalConcurrencyLimit.ActiveSlots)
	model.SlotDecayPerSecond = types.Float64Value(globalConcurrencyLimit.SlotDecayPerSecond)

	return nil
}

// Delete deletes a global concurrency limit.
func (r *GlobalConcurrencyLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GlobalConcurrencyLimits(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Global Concurrency Limit", err))

		return
	}

	err = client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Global Concurrency Limit", "delete", err))

		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads a global concurrency limit.
func (r *GlobalConcurrencyLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GlobalConcurrencyLimits(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Global Concurrency Limit", err))

		return
	}

	globalConcurrencyLimit, err := client.Read(ctx, state.ID.ValueString())
	if err != nil {
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Global Concurrency Limit", "read", err))

		return
	}

	copyGlobalConcurrencyLimitToModel(globalConcurrencyLimit, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates a global concurrency limit.
func (r *GlobalConcurrencyLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalConcurrencyLimitResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GlobalConcurrencyLimits(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Global Concurrency Limit", err))

		return
	}

	err = client.Update(ctx, plan.ID.ValueString(), api.GlobalConcurrencyLimitUpdate{
		Name:               plan.Name.ValueString(),
		Limit:              plan.Limit.ValueInt64(),
		Active:             plan.Active.ValueBool(),
		ActiveSlots:        plan.ActiveSlots.ValueInt64(),
		SlotDecayPerSecond: plan.SlotDecayPerSecond.ValueFloat64(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Global Concurrency Limit", "update", err))

		return
	}

	globalConcurrencyLimit, err := client.Read(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Global Concurrency Limit", "read", err))

		return
	}

	copyGlobalConcurrencyLimitToModel(globalConcurrencyLimit, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports a global concurrency limit.
func (r *GlobalConcurrencyLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
