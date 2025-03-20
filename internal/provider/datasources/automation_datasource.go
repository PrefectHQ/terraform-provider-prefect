package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &AutomationDataSource{}
	_ datasource.DataSourceWithConfigure = &AutomationDataSource{}
)

type AutomationDataSource struct {
	client api.PrefectClient
}

// NewAutomationDataSource returns a new AutomationDataSource.
//
//nolint:ireturn // required by Terraform API
func NewAutomationDataSource() datasource.DataSource {
	return &AutomationDataSource{}
}

// Metadata returns the data source type name.
func (d *AutomationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

// Schema defines the schema for the data source.
func (d *AutomationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Automation by its ID
<br>
For more information, see [automate overview](https://docs.prefect.io/v3/automate/index).
`,
			helpers.AllPlans...,
		),
		Attributes: AutomationSchema(),
	}
}

// Configure initializes runtime state for the data source.
func (d *AutomationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *AutomationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model AutomationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Automations(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Automation", err))

		return
	}

	automationID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Automation", err))

		return
	}

	automation, err := client.Get(ctx, automationID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Automation", "get", err))

		return
	}

	resp.Diagnostics.Append(mapAutomationAPIToTerraform(ctx, automation, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// mapAutomationAPIToTerraform copies an Automation API object => Terraform model.
// This helper is used when an API response needs to be translated for Terraform state.
func mapAutomationAPIToTerraform(ctx context.Context, apiAutomation *api.Automation, tfModel *AutomationDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Map base attributes
	tfModel.ID = customtypes.NewUUIDValue(apiAutomation.ID)
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

	// NOTE: .Trigger has to a pointer here, as the user won't be expected to set a
	// `trigger` attribute on the dataource - we only expect them to provide an
	// Automation ID to query. Therefore, we'll need to initialize Trigger here,
	// so that we can set the corresponding, nested attributes from the API call.
	tfModel.Trigger = &TriggerModel{}

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
			tfModel.Trigger.Compound = &CompoundTriggerAttributesModel{
				Within:   within,
				Triggers: compositeTriggers,
			}
		}

		if apiAutomation.Trigger.Type == utils.TriggerTypeSequence {
			tfModel.Trigger.Sequence = &SequenceTriggerAttributesModel{
				Within:   within,
				Triggers: compositeTriggers,
			}
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
		after, diagnostics := types.ListValueFrom(ctx, types.StringType, apiTrigger.After)
		diags.Append(diagnostics...)
		expect, diagnostics := types.ListValueFrom(ctx, types.StringType, apiTrigger.Expect)
		diags.Append(diagnostics...)
		forEach, diagnostics := types.ListValueFrom(ctx, types.StringType, apiTrigger.ForEach)
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
