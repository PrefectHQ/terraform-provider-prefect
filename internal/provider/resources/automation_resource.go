package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	automationClient, err := r.client.Automations(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	createAutomationRequest := api.AutomationUpsert{}
	resp.Diagnostics.Append(mapAutomationTerraformToAPI(ctx, &createAutomationRequest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createdAutomation, err := automationClient.Create(ctx, createAutomationRequest)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "create", err))

		return
	}

	resp.Diagnostics.Append(mapAutomationAPIToTerraform(ctx, createdAutomation, &plan)...)
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

	resp.Diagnostics.Append(mapAutomationAPIToTerraform(ctx, automation, &state)...)
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
	resp.Diagnostics.Append(mapAutomationTerraformToAPI(ctx, &updateAutomationRequest, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	resp.Diagnostics.Append(mapAutomationAPIToTerraform(ctx, updatedAutomation, &plan)...)
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
// Valid import IDs:
// <automation_id>
// <automation_id>,<workspace_id>.
func (r *AutomationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) > 2 || len(parts) == 0 {
		resp.Diagnostics.AddError(
			"Error importing Automation",
			"Import ID must be in the format of <automation_id> OR <automation_id>,<workspace_id>",
		)

		return
	}

	automationID := parts[0]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), automationID)...)

	if len(parts) == 2 && parts[1] != "" {
		workspaceID, err := uuid.Parse(parts[1])
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID.String())...)
	}

	// We need to set the trigger to an empty TriggerModel during import
	// to avoid null value errors (Value Conversion Errors) from the provider framework.
	// The `trigger` attribute is non-nullable, but because it is modeled
	// by pointing to another struct, the framework expects a value to exist.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("trigger"), TriggerModel{})...)
}

func (r *AutomationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("trigger").AtName(utils.TriggerTypeEvent),
			path.MatchRoot("trigger").AtName(utils.TriggerTypeMetric),
			path.MatchRoot("trigger").AtName(utils.TriggerTypeCompound),
			path.MatchRoot("trigger").AtName(utils.TriggerTypeSequence),
		),
	}
}

// mapAutomationAPIToTerraform copies an Automation API object => Terraform model.
// This helper is used when an API response needs to be translated for Terraform state.
func mapAutomationAPIToTerraform(ctx context.Context, apiAutomation *api.Automation, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Map base attributes
	tfModel.ID = types.StringValue(apiAutomation.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(apiAutomation.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(apiAutomation.Updated)
	tfModel.Name = types.StringValue(apiAutomation.Name)
	tfModel.Description = types.StringValue(apiAutomation.Description)
	tfModel.Enabled = types.BoolValue(apiAutomation.Enabled)

	// Map actions
	actions, diagnostics := mapActionsAPIToTerraform(apiAutomation.Actions)
	diags.Append(diagnostics...)
	actionsOnTrigger, diagnostics := mapActionsAPIToTerraform(apiAutomation.ActionsOnTrigger)
	diags.Append(diagnostics...)
	actionsOnResolve, diagnostics := mapActionsAPIToTerraform(apiAutomation.ActionsOnResolve)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return diags
	}
	tfModel.Actions = actions
	tfModel.ActionsOnTrigger = actionsOnTrigger
	tfModel.ActionsOnResolve = actionsOnResolve

	// Map trigger
	switch apiAutomation.Trigger.Type {
	case utils.TriggerTypeEvent, utils.TriggerTypeMetric:
		diags.Append(mapTriggerAPIToTerraform(ctx, &apiAutomation.Trigger, &tfModel.Trigger.ResourceTriggerModel)...)

	case utils.TriggerTypeCompound, utils.TriggerTypeSequence:

		// Iterate through API's composite triggers, map them to a Terraform model.
		compositeTriggers := make([]ResourceTriggerModel, 0)
		for i := range apiAutomation.Trigger.Triggers {
			apiTrigger := apiAutomation.Trigger.Triggers[i]
			resourceTriggerModel := ResourceTriggerModel{}
			diags.Append(mapTriggerAPIToTerraform(ctx, &apiTrigger, &resourceTriggerModel)...)
			compositeTriggers = append(compositeTriggers, resourceTriggerModel)
		}

		within := types.Float64PointerValue(apiAutomation.Trigger.Within)

		if apiAutomation.Trigger.Type == utils.TriggerTypeCompound {
			// NOTE: we'll carry forward whatever .Require value that
			// exists in State, so we don't need to re-set it here.
			tfModel.Trigger.Compound.Within = within
			tfModel.Trigger.Compound.Triggers = compositeTriggers
		}

		if apiAutomation.Trigger.Type == utils.TriggerTypeSequence {
			tfModel.Trigger.Sequence.Within = within
			tfModel.Trigger.Sequence.Triggers = compositeTriggers
		}
	default:
		diags.AddError("Invalid Trigger Type", fmt.Sprintf("Invalid trigger type: %s", apiAutomation.Trigger.Type))
	}

	return diags
}

// mapTriggerAPIToTerraform maps an `event` or `metric` trigger
// from an Automation API object => Terraform model.
// We map these separately, so we can re-use this helper for `compound` and `sequence` triggers.
func mapTriggerAPIToTerraform(ctx context.Context, apiTrigger *api.Trigger, tfTriggerModel *ResourceTriggerModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse Match and MatchRelated (JSON) regardless of type,
	// as they are common on Resource Trigger schemas.
	matchByteSlice, err := json.Marshal(apiTrigger.Match)
	if err != nil {
		diags.Append(helpers.SerializeDataErrorDiagnostic("match", "Automation trigger match", err))

		return diags
	}
	matchRelatedByteSlice, err := json.Marshal(apiTrigger.MatchRelated)
	if err != nil {
		diags.Append(helpers.SerializeDataErrorDiagnostic("match_related", "Automation trigger match related", err))

		return diags
	}

	switch apiTrigger.Type {
	case utils.TriggerTypeEvent:
		tfTriggerModel.Event = &EventTriggerModel{
			Posture:   types.StringPointerValue(apiTrigger.Posture),
			Threshold: types.Int64PointerValue(apiTrigger.Threshold),
			Within:    types.Float64PointerValue(apiTrigger.Within),
		}

		// Set Match and MatchRelated, which we parsed above.
		tfTriggerModel.Event.Match = jsontypes.NewNormalizedValue(string(matchByteSlice))
		tfTriggerModel.Event.MatchRelated = jsontypes.NewNormalizedValue(string(matchRelatedByteSlice))

		// Parse and set After, Expect, and ForEach (lists)
		after, diagnostics := types.SetValueFrom(ctx, types.StringType, apiTrigger.After)
		diags.Append(diagnostics...)
		expect, diagnostics := types.SetValueFrom(ctx, types.StringType, apiTrigger.Expect)
		diags.Append(diagnostics...)
		forEach, diagnostics := types.SetValueFrom(ctx, types.StringType, apiTrigger.ForEach)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		tfTriggerModel.Event.After = after
		tfTriggerModel.Event.Expect = expect
		tfTriggerModel.Event.ForEach = forEach

	case utils.TriggerTypeMetric:
		tfTriggerModel.Metric = &MetricTriggerModel{
			Metric: MetricQueryModel{
				Name:      types.StringValue(apiTrigger.Metric.Name),
				Threshold: types.Float64Value(apiTrigger.Metric.Threshold),
				Operator:  types.StringValue(apiTrigger.Metric.Operator),
				Range:     types.Float64Value(apiTrigger.Metric.Range),
				FiringFor: types.Float64Value(apiTrigger.Metric.FiringFor),
			},
		}

		// Set Match and MatchRelated, which we parsed above.
		tfTriggerModel.Metric.Match = jsontypes.NewNormalizedValue(string(matchByteSlice))
		tfTriggerModel.Metric.MatchRelated = jsontypes.NewNormalizedValue(string(matchRelatedByteSlice))

	default:
		diags.AddError("Invalid Trigger Type", fmt.Sprintf("Invalid trigger type: %s", apiTrigger.Type))
	}

	return diags
}

// mapActionsAPIToTerraform maps an Automation API Actions payload => Terraform model.
// This helper is used when an API response needs to be translated for Terraform state.
// NOTE: we are re-constructing the response list every time (rather than updating member items in-place),
// because there is no guarantee that the API returns the list in the same order as the Terraform model.
func mapActionsAPIToTerraform(apiActions []api.Action) ([]ActionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	tfActionsModel := make([]ActionModel, 0)

	for i := range apiActions {
		action := apiActions[i]

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
		} else {
			actionModel.Parameters = jsontypes.NewNormalizedValue("{}")
		}

		if action.JobVariables != nil {
			byteSlice, err := json.Marshal(action.JobVariables)
			if err != nil {
				diags.Append(helpers.SerializeDataErrorDiagnostic("job_variables", "Automation action job variables", err))

				return nil, diags
			}
			actionModel.JobVariables = jsontypes.NewNormalizedValue(string(byteSlice))
		} else {
			actionModel.JobVariables = jsontypes.NewNormalizedValue("{}")
		}

		tfActionsModel = append(tfActionsModel, actionModel)
	}

	return tfActionsModel, diags
}

// mapAutomationTerraformToAPI copies the Terraform model => AutomationUpsert request payload.
// This helper is used when a Terraform configuration needs to be translated for an API call.
func mapAutomationTerraformToAPI(ctx context.Context, apiAutomation *api.AutomationUpsert, tfModel *AutomationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Map base attributes
	apiAutomation.Name = tfModel.Name.ValueString()
	apiAutomation.Description = tfModel.Description.ValueString()
	apiAutomation.Enabled = tfModel.Enabled.ValueBool()

	// Map actions
	actions, diagnostics := mapActionsTerraformToAPI(tfModel.Actions)
	diags.Append(diagnostics...)
	actionsOnTrigger, diagnostics := mapActionsTerraformToAPI(tfModel.ActionsOnTrigger)
	diags.Append(diagnostics...)
	actionsOnResolve, diagnostics := mapActionsTerraformToAPI(tfModel.ActionsOnResolve)
	diags.Append(diagnostics...)
	if diags.HasError() {
		return diags
	}
	apiAutomation.Actions = actions
	apiAutomation.ActionsOnTrigger = actionsOnTrigger
	apiAutomation.ActionsOnResolve = actionsOnResolve

	// Map trigger
	switch {
	case tfModel.Trigger.Event != nil, tfModel.Trigger.Metric != nil:
		diags.Append(mapTriggerTerraformToAPI(ctx, &apiAutomation.Trigger, &tfModel.Trigger.ResourceTriggerModel)...)

	case tfModel.Trigger.Compound != nil:
		apiAutomation.Trigger.Type = utils.TriggerTypeCompound
		apiAutomation.Trigger.Within = tfModel.Trigger.Compound.Within.ValueFloat64Pointer()

		requireValue, diagnostics := getUnderlyingRequireValue(*tfModel.Trigger.Compound)
		diags.Append(diagnostics...)
		apiAutomation.Trigger.Require = &requireValue

		compositeTriggers := make([]api.Trigger, 0)
		for _, trigger := range tfModel.Trigger.Compound.Triggers {
			apiTrigger := api.Trigger{}
			diags.Append(mapTriggerTerraformToAPI(ctx, &apiTrigger, &trigger)...)
			compositeTriggers = append(compositeTriggers, apiTrigger)
		}
		apiAutomation.Trigger.Triggers = compositeTriggers

	case tfModel.Trigger.Sequence != nil:
		apiAutomation.Trigger.Type = utils.TriggerTypeSequence
		apiAutomation.Trigger.Within = tfModel.Trigger.Sequence.Within.ValueFloat64Pointer()

		compositeTriggers := make([]api.Trigger, 0)
		for _, trigger := range tfModel.Trigger.Sequence.Triggers {
			apiTrigger := api.Trigger{}
			diags.Append(mapTriggerTerraformToAPI(ctx, &apiTrigger, &trigger)...)
			compositeTriggers = append(compositeTriggers, apiTrigger)
		}
		apiAutomation.Trigger.Triggers = compositeTriggers

	default:
		diags.AddError("Invalid Trigger Type", "No valid trigger type specified")

		return diags
	}

	return diags
}

// mapTriggerTerraformToAPI maps an `event` or `metric` trigger
// from a Terraform model => AutomationUpsert request payload.
// We map these separately, so we can re-use this helper for `compound` and `sequence` triggers.
func mapTriggerTerraformToAPI(ctx context.Context, apiTrigger *api.Trigger, tfTriggerModel *ResourceTriggerModel) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case tfTriggerModel.Event != nil:
		var after, expect, forEach []string
		diags.Append(tfTriggerModel.Event.After.ElementsAs(ctx, &after, false)...)
		diags.Append(tfTriggerModel.Event.Expect.ElementsAs(ctx, &expect, false)...)
		diags.Append(tfTriggerModel.Event.ForEach.ElementsAs(ctx, &forEach, false)...)

		match, diagnostics := helpers.UnmarshalOptional(tfTriggerModel.Event.Match)
		diags.Append(diagnostics...)
		matchRelated, diagnostics := helpers.UnmarshalOptional(tfTriggerModel.Event.MatchRelated)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		*apiTrigger = api.Trigger{
			Type:         utils.TriggerTypeEvent,
			Posture:      tfTriggerModel.Event.Posture.ValueStringPointer(),
			Match:        match,
			MatchRelated: matchRelated,
			After:        after,
			Expect:       expect,
			ForEach:      forEach,
			Threshold:    tfTriggerModel.Event.Threshold.ValueInt64Pointer(),
			Within:       tfTriggerModel.Event.Within.ValueFloat64Pointer(),
		}
	case tfTriggerModel.Metric != nil:
		match, diagnostics := helpers.UnmarshalOptional(tfTriggerModel.Metric.Match)
		diags.Append(diagnostics...)
		matchRelated, diagnostics := helpers.UnmarshalOptional(tfTriggerModel.Metric.MatchRelated)
		diags.Append(diagnostics...)

		if diags.HasError() {
			return diags
		}

		*apiTrigger = api.Trigger{
			Type:         utils.TriggerTypeMetric,
			Match:        match,
			MatchRelated: matchRelated,
			Metric: &api.MetricTriggerQuery{
				Name:      tfTriggerModel.Metric.Metric.Name.ValueString(),
				Threshold: tfTriggerModel.Metric.Metric.Threshold.ValueFloat64(),
				Operator:  tfTriggerModel.Metric.Metric.Operator.ValueString(),
				Range:     tfTriggerModel.Metric.Metric.Range.ValueFloat64(),
				FiringFor: tfTriggerModel.Metric.Metric.FiringFor.ValueFloat64(),
			},
		}

	default:
		diags.AddError("Invalid Trigger Type", "No valid trigger type specified")
	}

	return diags
}

// mapActionsTerraformToAPI creates an Actions payload for an AutomationUpsert request.
// This helper is used when constructing the overall AutomationUpsert request payload
// from a given Terraform model.
func mapActionsTerraformToAPI(tfActions []ActionModel) ([]api.Action, diag.Diagnostics) {
	var diags diag.Diagnostics

	actions := make([]api.Action, 0)

	for i := range tfActions {
		tfAction := tfActions[i]
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
		parameters, diagnostics := helpers.UnmarshalOptional(tfAction.Parameters)
		diags.Append(diagnostics...)
		jobVariables, diagnostics := helpers.UnmarshalOptional(tfAction.JobVariables)
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

// getUnderlyingRequireValue extracts the underlying value from a CompositeTriggerAttributesModel's Require field.
// This helper is used when constructing the overall AutomationUpsert request payload
// from a given Terraform model, as the .Require attribute is a types.DynamicValue.
func getUnderlyingRequireValue(plan CompoundTriggerAttributesModel) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var value interface{}

	switch underlyingValue := plan.Require.UnderlyingValue().(type) {
	case types.String:
		value = underlyingValue.ValueString()
	case types.Number:
		var err error
		value, err = strconv.ParseInt(underlyingValue.String(), 10, 64)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"unable to convert number to int64",
				fmt.Sprintf("number: %v, error: %v", value, err),
			))
		}
	default:
		diags.AddError("Invalid Type", fmt.Sprintf("Invalid type for compound trigger's require value: %T", underlyingValue))
	}

	return value, diags
}
