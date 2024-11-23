package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&AutomationResource{})
	_ = resource.ResourceWithImportState(&AutomationResource{})
	_ = resource.ResourceWithConfigValidators(&AutomationResource{})
)

// AutomationResource contains state for the resource.
type AutomationResource struct {
	client api.PrefectClient
}

// NewAutomationResource returns a new AutomationResource.
//
//nolint:ireturn // required by Terraform API
func NewAutomationResource() resource.Resource {
	return &AutomationResource{}
}

// Metadata returns the resource type name.
func (r *AutomationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

// Configure initializes runtime state for the resource.
func (r *AutomationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected api.PrefectClient, got: "+fmt.Sprintf("%T", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *AutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `automations` represents a Prefect Automation.",
		Version:     0,
		Attributes:  AutomationSchema(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *AutomationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AutomationResourceModel

	tflog.Info(ctx, "WE ARE HERE")

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("plan: %+v", plan))

	automationClient, err := r.client.Automations(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	trigger, diags := validateAndCreateTriggerPayload(ctx, plan)
	resp.Diagnostics.Append(diags...)

	actions, diags := createActionsPayload(plan.Actions)
	resp.Diagnostics.Append(diags...)

	actionsOnTrigger, diags := createActionsPayload(plan.ActionsOnTrigger)
	resp.Diagnostics.Append(diags...)

	actionsOnResolve, diags := createActionsPayload(plan.ActionsOnResolve)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createAutomationRequest := api.AutomationUpsert{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		Enabled:          plan.Enabled.ValueBool(),
		Trigger:          *trigger,
		Actions:          actions,
		ActionsOnTrigger: actionsOnTrigger,
		ActionsOnResolve: actionsOnResolve,
	}

	createdAutomation, err := automationClient.Create(ctx, createAutomationRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "create", err))

		return
	}

	// TODO: move to translation helper function
	plan.ID = types.StringValue(createdAutomation.ID.String())
	plan.Created = customtypes.NewTimestampPointerValue(createdAutomation.Created)
	plan.Updated = customtypes.NewTimestampPointerValue(createdAutomation.Updated)
	plan.Name = types.StringValue(createdAutomation.Name)
	plan.Description = types.StringValue(createdAutomation.Description)
	plan.Enabled = types.BoolValue(createdAutomation.Enabled)

	// TODO: do we need to re-set the trigger here?

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *AutomationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AutomationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AutomationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AutomationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	automationClient, err := r.client.Automations(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	automationID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Automation", err))

		return
	}

	err = automationClient.Delete(ctx, automationID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *AutomationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

func (r *AutomationResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("trigger").AtName("event"),
			path.MatchRoot("trigger").AtName("metric"),
			path.MatchRoot("trigger").AtName("compound"),
			path.MatchRoot("trigger").AtName("sequence"),
		),
	}
}

func validateAndCreateTriggerPayload(ctx context.Context, plan AutomationResourceModel) (*api.Trigger, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case plan.Trigger.Event != nil:
		var after, expect, forEach []string
		diags.Append(plan.Trigger.Event.After.ElementsAs(ctx, &after, false)...)
		diags.Append(plan.Trigger.Event.Expect.ElementsAs(ctx, &expect, false)...)
		diags.Append(plan.Trigger.Event.ForEach.ElementsAs(ctx, &forEach, false)...)

		match, diagnostics := helpers.SafeUnmarshal(plan.Trigger.Event.Match)
		diags.Append(diagnostics...)
		matchRelated, diagnostics := helpers.SafeUnmarshal(plan.Trigger.Event.MatchRelated)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return nil, diags
		}

		return &api.Trigger{
			Type:         "event",
			Posture:      plan.Trigger.Event.Posture.ValueStringPointer(),
			Match:        match,
			MatchRelated: matchRelated,
			After:        after,
			Expect:       expect,
			ForEach:      forEach,
			Threshold:    plan.Trigger.Event.Threshold.ValueInt64Pointer(),
			Within:       plan.Trigger.Event.Within.ValueFloat64Pointer(),
		}, nil

	// case plan.Trigger.Metric != nil:
	// 	return api.Trigger{Metric: plan.Trigger.Metric}, nil
	// case plan.Trigger.Compound != nil:
	// 	return api.Trigger{Compound: plan.Trigger.Compound}, nil
	// case plan.Trigger.Sequence != nil:
	// 	return api.Trigger{Sequence: plan.Trigger.Sequence}, nil
	default:
		diags.AddError("Invalid Trigger Type", "No valid trigger type specified")
		return nil, diags
	}
}

func createActionsPayload(tfActions []ActionModel) ([]api.Action, diag.Diagnostics) {
	actions := make([]api.Action, 0)

	for _, tfAction := range tfActions {
		apiAction := api.Action{}

		apiAction.Type = tfAction.Type.ValueString()
		apiAction.Source = tfAction.Source.ValueStringPointer()
		apiAction.AutomationID = tfAction.AutomationID.ValueUUIDPointer()
		apiAction.BlockDocumentID = tfAction.BlockDocumentID.ValueUUIDPointer()
		apiAction.DeploymentID = tfAction.DeploymentID.ValueUUIDPointer()
		apiAction.WorkPoolID = tfAction.WorkPoolID.ValueUUIDPointer()
		apiAction.WorkQueueID = tfAction.WorkQueueID.ValueUUIDPointer()
		apiAction.Subject = tfAction.Subject.ValueStringPointer()
		apiAction.Body = tfAction.Body.ValueStringPointer()
		apiAction.Payload = tfAction.Payload.ValueStringPointer()
		apiAction.Name = tfAction.Name.ValueStringPointer()
		apiAction.State = tfAction.State.ValueStringPointer()
		apiAction.Message = tfAction.Message.ValueStringPointer()

		parameters, diags := helpers.SafeUnmarshal(tfAction.Parameters)
		if diags.HasError() {
			return nil, diags
		}
		apiAction.Parameters = parameters

		jobVariables, diags := helpers.SafeUnmarshal(tfAction.JobVariables)
		if diags.HasError() {
			return nil, diags
		}
		apiAction.JobVariables = jobVariables

		actions = append(actions, apiAction)
	}

	return actions, nil
}
