// TODO: rename helpers for consistency
// TODO: ensure helper argument names make sense
package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	automationClient, err := r.client.Automations(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	createAutomationRequest := api.AutomationUpsert{}
	resp.Diagnostics.Append(copyModelToAutomationRequest(ctx, &createAutomationRequest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdAutomation, err := automationClient.Create(ctx, createAutomationRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "create", err))

		return
	}

	resp.Diagnostics.Append(copyAutomationToModel(ctx, createdAutomation, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *AutomationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AutomationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Automations(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	automationID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Automation", err))

		return
	}

	automation, err := client.Get(ctx, automationID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "get", err))

		return
	}

	resp.Diagnostics.Append(copyAutomationToModel(ctx, automation, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AutomationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AutomationResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	automationClient, err := r.client.Automations(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	updateAutomationRequest := api.AutomationUpsert{}
	resp.Diagnostics.Append(copyModelToAutomationRequest(ctx, &updateAutomationRequest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, fmt.Sprintf("planID: %+v", plan.ID.ValueString()))

	automationID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Automation", err))

		return
	}

	err = automationClient.Update(ctx, automationID, updateAutomationRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "update", err))

		return
	}

	updatedAutomation, err := automationClient.Get(ctx, automationID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "get", err))

		return
	}

	resp.Diagnostics.Append(copyAutomationToModel(ctx, updatedAutomation, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
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

// copyAutomationToModel copies an Automation response payload => Terraform model.
// This helper is used when an API response needs to be translated for Terraform state.
func copyAutomationToModel(ctx context.Context, automation *api.Automation, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Map base attributes
	tfModel.ID = types.StringValue(automation.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(automation.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(automation.Updated)
	tfModel.Name = types.StringValue(automation.Name)
	tfModel.Description = types.StringValue(automation.Description)
	tfModel.Enabled = types.BoolValue(automation.Enabled)

	// Map actions
	actions, diagnostics := createActionsForModel(automation.Actions)
	diags.Append(diagnostics...)
	actionsOnTrigger, diagnostics := createActionsForModel(automation.ActionsOnTrigger)
	diags.Append(diagnostics...)
	actionsOnResolve, diagnostics := createActionsForModel(automation.ActionsOnResolve)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return diags
	}
	tfModel.Actions = actions
	tfModel.ActionsOnTrigger = actionsOnTrigger
	tfModel.ActionsOnResolve = actionsOnResolve

	// Map trigger
	switch automation.Trigger.Type {
	case "event":
		diags.Append(mapResourceTriggerToModel(ctx, automation, tfModel)...)
	case "metric":
		diags.Append(mapResourceTriggerToModel(ctx, automation, tfModel)...)
	case "compound":
	case "sequence":
	default:
		diags.AddError("Invalid Trigger Type", fmt.Sprintf("Invalid trigger type: %s", automation.Trigger.Type))
	}

	return diags
}

// mapResourceTriggerToModel maps an `event` or `metric` trigger
// from an Automation response payload => Terraform model.
// We map these separately, so we can re-use this helper for `compound` and `sequence` triggers.
func mapResourceTriggerToModel(ctx context.Context, automation *api.Automation, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse Match and MatchRelated (JSON) regardless of type,
	// as they are common on Resource Trigger schemas.
	matchByteSlice, err := json.Marshal(automation.Trigger.Match)
	if err != nil {
		diags.Append(helpers.SerializeDataErrorDiagnostic("match", "Automation trigger match", err))

		return diags
	}
	matchRelatedByteSlice, err := json.Marshal(automation.Trigger.MatchRelated)
	if err != nil {
		diags.Append(helpers.SerializeDataErrorDiagnostic("match_related", "Automation trigger match related", err))

		return diags
	}

	switch automation.Trigger.Type {
	case "event":
		tfModel.Trigger.Event = &EventTriggerModel{
			Posture:   types.StringValue(*automation.Trigger.Posture),
			Threshold: types.Int64Value(*automation.Trigger.Threshold),
			Within:    types.Float64Value(*automation.Trigger.Within),
		}

		// Set Match and MatchRelated, which we parsed above.
		tfModel.Trigger.Event.Match = jsontypes.NewNormalizedValue(string(matchByteSlice))
		tfModel.Trigger.Event.MatchRelated = jsontypes.NewNormalizedValue(string(matchRelatedByteSlice))

		// Parse and set After, Expect, and ForEach (lists)
		after, diagnostics := types.ListValueFrom(ctx, types.StringType, automation.Trigger.After)
		diags.Append(diagnostics...)
		expect, diagnostics := types.ListValueFrom(ctx, types.StringType, automation.Trigger.Expect)
		diags.Append(diagnostics...)
		forEach, diagnostics := types.ListValueFrom(ctx, types.StringType, automation.Trigger.ForEach)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		tfModel.Trigger.Event.After = after
		tfModel.Trigger.Event.Expect = expect
		tfModel.Trigger.Event.ForEach = forEach

	case "metric":
		tfModel.Trigger.Metric = &MetricTriggerModel{
			Metric: MetricQueryModel{
				Name:      types.StringValue(automation.Trigger.Metric.Name),
				Threshold: types.Float64Value(automation.Trigger.Metric.Threshold),
				Operator:  types.StringValue(automation.Trigger.Metric.Operator),
				Range:     types.Float64Value(automation.Trigger.Metric.Range),
				FiringFor: types.Float64Value(automation.Trigger.Metric.FiringFor),
			},
		}

		// Set Match and MatchRelated, which we parsed above.
		tfModel.Trigger.Metric.Match = jsontypes.NewNormalizedValue(string(matchByteSlice))
		tfModel.Trigger.Metric.MatchRelated = jsontypes.NewNormalizedValue(string(matchRelatedByteSlice))

	default:
		diags.AddError("Invalid Trigger Type", fmt.Sprintf("Invalid trigger type: %s", automation.Trigger.Type))
	}

	return diags
}

func createActionsForModel(apiActions []api.Action) ([]ActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	actions := make([]ActionModel, 0)

	for _, action := range apiActions {
		actionModel := ActionModel{}
		actionModel.Type = types.StringValue(action.Type)
		actionModel.Source = types.StringPointerValue(action.Source)
		actionModel.AutomationID = customtypes.NewUUIDPointerValue(action.AutomationID)
		actionModel.BlockDocumentID = customtypes.NewUUIDPointerValue(action.BlockDocumentID)
		actionModel.DeploymentID = customtypes.NewUUIDPointerValue(action.DeploymentID)
		actionModel.WorkPoolID = customtypes.NewUUIDPointerValue(action.WorkPoolID)
		actionModel.WorkQueueID = customtypes.NewUUIDPointerValue(action.WorkQueueID)
		actionModel.Subject = types.StringPointerValue(action.Subject)
		actionModel.Body = types.StringPointerValue(action.Body)
		actionModel.Payload = types.StringPointerValue(action.Payload)
		actionModel.Name = types.StringPointerValue(action.Name)
		actionModel.State = types.StringPointerValue(action.State)
		actionModel.Message = types.StringPointerValue(action.Message)

		// Only set parameters and job variables if they are set in the API.
		// Otherwise, the string `"null"` is set to the Terraform model, which will
		// create an inconsistent result error if no value is set in HCL.
		if action.Parameters != nil {
			byteSlice, err := json.Marshal(action.Parameters)
			if err != nil {
				diags.Append(helpers.SerializeDataErrorDiagnostic("parameters", "Automation action parameters", err))

				return nil, diags
			}
			actionModel.Parameters = jsontypes.NewNormalizedValue(string(byteSlice))
		}
		if action.JobVariables != nil {
			byteSlice, err := json.Marshal(action.JobVariables)
			if err != nil {
				diags.Append(helpers.SerializeDataErrorDiagnostic("job_variables", "Automation action job variables", err))

				return nil, diags
			}
			actionModel.JobVariables = jsontypes.NewNormalizedValue(string(byteSlice))
		}

		actions = append(actions, actionModel)
	}

	return actions, diags
}

// copyModelToAutomationRequest copies the Terraform model => AutomationUpsert request payload.
// This helper is used when a Terraform configuration needs to be translated for an API call.
func copyModelToAutomationRequest(ctx context.Context, automationRequest *api.AutomationUpsert, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Map base attributes
	automationRequest.Name = tfModel.Name.ValueString()
	automationRequest.Description = tfModel.Description.ValueString()
	automationRequest.Enabled = tfModel.Enabled.ValueBool()

	// Map actions
	actions, diagnostics := createActionsForAutomationRequest(tfModel.Actions)
	diags.Append(diagnostics...)
	actionsOnTrigger, diagnostics := createActionsForAutomationRequest(tfModel.ActionsOnTrigger)
	diags.Append(diagnostics...)
	actionsOnResolve, diagnostics := createActionsForAutomationRequest(tfModel.ActionsOnResolve)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return diags
	}
	automationRequest.Actions = actions
	automationRequest.ActionsOnTrigger = actionsOnTrigger
	automationRequest.ActionsOnResolve = actionsOnResolve

	// Map trigger
	switch {
	case tfModel.Trigger.Event != nil:
		diags.Append(mapResourceTriggerToAutomationRequest(ctx, automationRequest, tfModel)...)
	case tfModel.Trigger.Metric != nil:
		diags.Append(mapResourceTriggerToAutomationRequest(ctx, automationRequest, tfModel)...)
	// case tfModel.Trigger.Compound != nil:
	// case tfModel.Trigger.Sequence != nil:
	default:
		diags.AddError("Invalid Trigger Type", "No valid trigger type specified")
		return diags
	}

	return diags
}

// mapResourceTriggerToAutomationRequest maps an `event` or `metric` trigger
// from a Terraform model => AutomationUpsert request payload.
// We map these separately, so we can re-use this helper for `compound` and `sequence` triggers.
func mapResourceTriggerToAutomationRequest(ctx context.Context, automationRequest *api.AutomationUpsert, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case tfModel.Trigger.Event != nil:
		var after, expect, forEach []string
		diags.Append(tfModel.Trigger.Event.After.ElementsAs(ctx, &after, false)...)
		diags.Append(tfModel.Trigger.Event.Expect.ElementsAs(ctx, &expect, false)...)
		diags.Append(tfModel.Trigger.Event.ForEach.ElementsAs(ctx, &forEach, false)...)

		match, diagnostics := helpers.SafeUnmarshal(tfModel.Trigger.Event.Match)
		diags.Append(diagnostics...)
		matchRelated, diagnostics := helpers.SafeUnmarshal(tfModel.Trigger.Event.MatchRelated)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		automationRequest.Trigger = api.Trigger{
			Type:         "event",
			Posture:      tfModel.Trigger.Event.Posture.ValueStringPointer(),
			Match:        match,
			MatchRelated: matchRelated,
			After:        after,
			Expect:       expect,
			ForEach:      forEach,
			Threshold:    tfModel.Trigger.Event.Threshold.ValueInt64Pointer(),
			Within:       tfModel.Trigger.Event.Within.ValueFloat64Pointer(),
		}
	case tfModel.Trigger.Metric != nil:
		match, diagnostics := helpers.SafeUnmarshal(tfModel.Trigger.Metric.Match)
		diags.Append(diagnostics...)
		matchRelated, diagnostics := helpers.SafeUnmarshal(tfModel.Trigger.Metric.MatchRelated)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		automationRequest.Trigger = api.Trigger{
			Type:         "metric",
			Match:        match,
			MatchRelated: matchRelated,
			Metric: &api.MetricTriggerQuery{
				Name:      tfModel.Trigger.Metric.Metric.Name.ValueString(),
				Threshold: tfModel.Trigger.Metric.Metric.Threshold.ValueFloat64(),
				Operator:  tfModel.Trigger.Metric.Metric.Operator.ValueString(),
				Range:     tfModel.Trigger.Metric.Metric.Range.ValueFloat64(),
				FiringFor: tfModel.Trigger.Metric.Metric.FiringFor.ValueFloat64(),
			},
		}

	default:
		diags.AddError("Invalid Trigger Type", "No valid trigger type specified")
	}

	return diags
}

// createActionsForAutomationRequest creates an Actions payload for an AutomationUpsert request.
// This helper is used when constructing the overall AutomationUpsert request payload
// from a given Terraform model.
func createActionsForAutomationRequest(tfActions []ActionModel) ([]api.Action, diag.Diagnostics) {
	var diags diag.Diagnostics

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

		// Parse and set parameters + job variables (JSON)
		parameters, diagnostics := helpers.SafeUnmarshal(tfAction.Parameters)
		diags.Append(diagnostics...)
		jobVariables, diagnostics := helpers.SafeUnmarshal(tfAction.JobVariables)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return nil, diags
		}

		apiAction.Parameters = parameters
		apiAction.JobVariables = jobVariables

		actions = append(actions, apiAction)
	}

	return actions, diags
}
