package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&DeploymentScheduleResource{})

type DeploymentScheduleResource struct {
	client api.PrefectClient
}

type DeploymentScheduleResourceModel struct {
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	DeploymentID customtypes.UUIDValue `tfsdk:"deployment_id"`

	Active           types.Bool    `tfsdk:"active"`
	MaxScheduledRuns types.Float32 `tfsdk:"max_scheduled_runs"`

	// Cloud-only
	MaxActiveRuns types.Float32 `tfsdk:"max_active_runs"`
	Catchup       types.Bool    `tfsdk:"catchup"`

	// All schedule kinds specify an interval.
	Timezone types.String `tfsdk:"timezone"`

	// Schedule kind: interval
	Interval   types.Float32 `tfsdk:"interval"`
	AnchorDate types.String  `tfsdk:"anchor_date"`

	// Schedule kind: cron
	Cron  types.String `tfsdk:"cron"`
	DayOr types.Bool   `tfsdk:"day_or"`

	// Schedule kind: rrule
	RRule types.String `tfsdk:"rrule"`
}

// NewDeploymentScheduleResource returns a new DeploymentScheduleResource.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentScheduleResource() resource.Resource {
	return &DeploymentScheduleResource{}
}

// Metadata returns the resource type name.
func (r *DeploymentScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_schedule"
}

// Configure initializes runtime state for the resource.
func (r *DeploymentScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `deployment_schedule` represents a schedule for a deployment.",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Deployment Schedule ID (UUID)",
				Optional:    true,
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
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Account ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				Description: "Workspace ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"deployment_id": schema.StringAttribute{
				Required:    true,
				Description: "Deployment ID (UUID)",
				CustomType:  customtypes.UUIDType{},
			},
			"active": schema.BoolAttribute{
				Description: "Whether or not the schedule is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"max_scheduled_runs": schema.Float32Attribute{
				Description: "The maximum number of scheduled runs for the schedule.",
				Optional:    true,
				Computed:    true,
			},
			"max_active_runs": schema.Float32Attribute{
				Description: "(Cloud only) The maximum number of active runs for the schedule.",
				Optional:    true,
				Computed:    true,
			},
			"catchup": schema.BoolAttribute{
				Description: "(Cloud only) Whether or not a worker should catch up on Late runs for the schedule.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// Timezone is a common field for all schedule kinds.
			"timezone": schema.StringAttribute{
				Description: "The timezone of the schedule.",
				Optional:    true,
				Computed:    true,
			},
			// Schedule kind: interval
			"interval": schema.Float32Attribute{
				Description: "The interval of the schedule.",
				Optional:    true,
				Computed:    true,
			},
			"anchor_date": schema.StringAttribute{
				Description: "The anchor date of the schedule.",
				Optional:    true,
				Computed:    true,
			},
			// Schedule kind: cron
			"cron": schema.StringAttribute{
				Description: "The cron expression of the schedule.",
				Optional:    true,
				Computed:    true,
			},
			"day_or": schema.BoolAttribute{
				Description: "Control croniter behavior for handling day and day_of_week entries.",
				Optional:    true,
				Computed:    true,
			},
			// Schedule kind: rrule
			"rrule": schema.StringAttribute{
				Description: "The rrule expression of the schedule.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *DeploymentScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeploymentScheduleResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.DeploymentSchedule(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Schedule", err))

		return
	}

	cfgCreate := []api.DeploymentSchedulePayload{
		{
			Active:           plan.Active.ValueBool(),
			MaxScheduledRuns: plan.MaxScheduledRuns.ValueFloat32(),
			MaxActiveRuns:    plan.MaxActiveRuns.ValueFloat32(),
			Catchup:          plan.Catchup.ValueBool(),
			Schedule: api.Schedule{
				AnchorDate: plan.AnchorDate.ValueString(),
				Cron:       plan.Cron.ValueString(),
				DayOr:      plan.DayOr.ValueBool(),
				Interval:   plan.Interval.ValueFloat32(),
				RRule:      plan.RRule.ValueString(),
				Timezone:   plan.Timezone.ValueString(),
			},
		},
	}

	schedules, err := client.Create(ctx, plan.DeploymentID.ValueUUID(), cfgCreate)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Schedule", "create", err))

		return
	}

	for _, schedule := range schedules {
		copyScheduleModelToResourceModel(schedule, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource and sets the Terraform state.
func (r *DeploymentScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeploymentScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.DeploymentSchedule(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Schedule", err))

		return
	}

	schedules, err := client.Read(ctx, state.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Schedule", "read", err))

		return
	}

	copyScheduleModelToResourceModel(schedules[0], &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the Terraform state.
func (r *DeploymentScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeploymentScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.DeploymentSchedule(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Schedule", err))

		return
	}

	cfgUpdate := []api.DeploymentSchedulePayload{
		{
			Active:           plan.Active.ValueBool(),
			MaxScheduledRuns: plan.MaxScheduledRuns.ValueFloat32(),
			MaxActiveRuns:    plan.MaxActiveRuns.ValueFloat32(),
			Schedule: api.Schedule{
				AnchorDate: plan.AnchorDate.ValueString(),
				Cron:       plan.Cron.ValueString(),
				DayOr:      plan.DayOr.ValueBool(),
				Interval:   plan.Interval.ValueFloat32(),
				RRule:      plan.RRule.ValueString(),
				Timezone:   plan.Timezone.ValueString(),
			},
		},
	}

	err = client.Update(ctx, plan.DeploymentID.ValueUUID(), plan.ID.ValueUUID(), cfgUpdate)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Schedule", "update", err))

		return
	}

	schedules, err := client.Read(ctx, plan.DeploymentID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Schedule", "read", err))

		return
	}

	// this isn't ideal. maybe go back
	copyScheduleModelToResourceModel(schedules[0], &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DeploymentScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeploymentScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.DeploymentSchedule(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment Schedule", err))

		return
	}

	err = client.Delete(ctx, state.DeploymentID.ValueUUID(), state.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment Schedule", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func copyScheduleModelToResourceModel(schedule *api.DeploymentSchedule, model *DeploymentScheduleResourceModel) {
	model.ID = customtypes.NewUUIDValue(schedule.ID)
	model.Created = customtypes.NewTimestampPointerValue(schedule.Created)
	model.Updated = customtypes.NewTimestampPointerValue(schedule.Updated)

	model.DeploymentID = customtypes.NewUUIDValue(schedule.DeploymentID)

	model.Active = types.BoolValue(schedule.Active)
	model.MaxActiveRuns = types.Float32Value(schedule.MaxActiveRuns)

	model.Catchup = types.BoolValue(schedule.Catchup)
	model.MaxScheduledRuns = types.Float32Value(schedule.MaxScheduledRuns)

	model.Timezone = types.StringPointerValue(&schedule.Schedule.Timezone)
	model.Interval = types.Float32Value(schedule.Schedule.Interval)
	model.AnchorDate = types.StringPointerValue(&schedule.Schedule.AnchorDate)
	model.Cron = types.StringPointerValue(&schedule.Schedule.Cron)
	model.DayOr = types.BoolValue(schedule.Schedule.DayOr)
	model.RRule = types.StringPointerValue(&schedule.Schedule.RRule)
}
