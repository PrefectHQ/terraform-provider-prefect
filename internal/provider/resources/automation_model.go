package resources

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// AutomationResourceModel defines the Terraform resource model.
type AutomationResourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Trigger     TriggerModel `tfsdk:"trigger"`

	Actions          []ActionModel `tfsdk:"actions"`
	ActionsOnTrigger []ActionModel `tfsdk:"actions_on_trigger"`
	ActionsOnResolve []ActionModel `tfsdk:"actions_on_resolve"`
}

// ResourceTriggerModel comprises the event and metric trigger models.
type ResourceTriggerModel struct {
	Event  *EventTriggerModel  `tfsdk:"event"`
	Metric *MetricTriggerModel `tfsdk:"metric"`
}

// TriggerModel comprises all trigger types, including compound and sequence.
type TriggerModel struct {
	ResourceTriggerModel
	Compound *CompositeTriggerAttributesModel `tfsdk:"compound"`
	Sequence *CompositeTriggerAttributesModel `tfsdk:"sequence"`
}

// EventTriggerModel represents an event-based trigger
type EventTriggerModel struct {
	Posture      types.String         `tfsdk:"posture"`
	Match        jsontypes.Normalized `tfsdk:"match"`
	MatchRelated jsontypes.Normalized `tfsdk:"match_related"`
	After        types.List           `tfsdk:"after"`
	Expect       types.List           `tfsdk:"expect"`
	ForEach      types.List           `tfsdk:"for_each"`
	Threshold    types.Int64          `tfsdk:"threshold"`
	Within       types.Float64        `tfsdk:"within"`
}

// MetricTriggerModel represents a metric-based trigger
type MetricTriggerModel struct {
	Match        jsontypes.Normalized `tfsdk:"match"`
	MatchRelated jsontypes.Normalized `tfsdk:"match_related"`
	Metric       MetricQueryModel     `tfsdk:"metric"`
}

// MetricQueryModel represents the metric query configuration
type MetricQueryModel struct {
	Name      types.String  `tfsdk:"name"`
	Operator  types.String  `tfsdk:"operator"`
	Threshold types.Float64 `tfsdk:"threshold"`
	Range     types.Float64 `tfsdk:"range"`
	FiringFor types.Float64 `tfsdk:"firing_for"`
}

// CompositeTriggerAttributesModel represents the shared
// attributes of a compound or sequence trigger.
type CompositeTriggerAttributesModel struct {
	Triggers []ResourceTriggerModel `tfsdk:"triggers"`
	Within   types.Float64          `tfsdk:"within"`
	Require  types.Dynamic          `tfsdk:"require"` // only exists on compound triggers
}

// ActionModel represents a single action in an automation
type ActionModel struct {
	// On all actions
	Type types.String `tfsdk:"type"`

	// On Deployment, Work Pool, Work Queue, and Automation actions
	Source types.String `tfsdk:"source"`

	// On Automation actions
	AutomationID customtypes.UUIDValue `tfsdk:"automation_id"`

	// On Webhook and Notification actions
	BlockDocumentID customtypes.UUIDValue `tfsdk:"block_document_id"`

	// On Deployment actions
	DeploymentID customtypes.UUIDValue `tfsdk:"deployment_id"`

	// On Work Pool actions
	WorkPoolID customtypes.UUIDValue `tfsdk:"work_pool_id"`

	// On Work Queue actions
	WorkQueueID customtypes.UUIDValue `tfsdk:"work_queue_id"`

	// On Run Deployment action
	Parameters   jsontypes.Normalized `tfsdk:"parameters"`
	JobVariables jsontypes.Normalized `tfsdk:"job_variables"`

	// On Send Notification action
	Subject types.String `tfsdk:"subject"`
	Body    types.String `tfsdk:"body"`

	// On Call Webhook action
	Payload types.String `tfsdk:"payload"`

	// On Flow Run State Change action
	Name    types.String `tfsdk:"name"`
	State   types.String `tfsdk:"state"`
	Message types.String `tfsdk:"message"`
}
