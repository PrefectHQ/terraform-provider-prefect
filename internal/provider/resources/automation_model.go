package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// AutomationsResourceModel defines the Terraform resource model.
type AutomationsResourceModel struct {
	ID        types.String               `tfsdk:"id"`
	Created   customtypes.TimestampValue `tfsdk:"created"`
	Updated   customtypes.TimestampValue `tfsdk:"updated"`
	AccountID customtypes.UUIDValue      `tfsdk:"account_id"`

	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	Enabled          types.Bool    `tfsdk:"enabled"`
	Trigger          TriggerModel  `tfsdk:"trigger"`
	Actions          []ActionModel `tfsdk:"actions"`
	ActionsOnTrigger []ActionModel `tfsdk:"actions_on_trigger"`
	ActionsOnResolve []ActionModel `tfsdk:"actions_on_resolve"`
}

// TriggerModel represents the top-level trigger configuration
type TriggerModel struct {
	Event    *EventTriggerModel    `tfsdk:"event"`
	Metric   *MetricTriggerModel   `tfsdk:"metric"`
	Compound *CompoundTriggerModel `tfsdk:"compound"`
	Sequence *SequenceTriggerModel `tfsdk:"sequence"`
}

// EventTriggerModel represents an event-based trigger
type EventTriggerModel struct {
	Posture      types.String `tfsdk:"posture"`
	Match        types.Map    `tfsdk:"match"`
	MatchRelated types.Map    `tfsdk:"match_related"`
	After        types.Set    `tfsdk:"after"`
	Expect       types.Set    `tfsdk:"expect"`
	ForEach      types.Set    `tfsdk:"for_each"`
	Threshold    types.Int64  `tfsdk:"threshold"`
	Within       types.Int64  `tfsdk:"within"`
}

// MetricTriggerModel represents a metric-based trigger
type MetricTriggerModel struct {
	Posture types.String     `tfsdk:"posture"`
	Metric  MetricQueryModel `tfsdk:"metric"`
}

// MetricQueryModel represents the metric query configuration
type MetricQueryModel struct {
	Name      types.String  `tfsdk:"name"`
	Operator  types.String  `tfsdk:"operator"`
	Threshold types.Float64 `tfsdk:"threshold"`
	Range     types.Int64   `tfsdk:"range"`
}

// CompoundTriggerModel represents a compound trigger
type CompoundTriggerModel struct {
	Require  types.String   `tfsdk:"require"`
	Triggers []TriggerModel `tfsdk:"triggers"`
}

// SequenceTriggerModel represents a sequence trigger
type SequenceTriggerModel struct {
	Triggers []TriggerModel `tfsdk:"triggers"`
}

// ActionModel represents a single action in an automation
type ActionModel struct {
	RunDeployment      *RunDeploymentAction    `tfsdk:"run-deployment"`
	SendNotification   *SendNotificationAction `tfsdk:"send-notification"`
	CallWebhook        *CallWebhookAction      `tfsdk:"call-webhook"`
	PauseDeployment    *DeploymentAction       `tfsdk:"pause-deployment"`
	ResumeDeployment   *DeploymentAction       `tfsdk:"resume-deployment"`
	CancelFlowRun      *EmptyAction            `tfsdk:"cancel-flow-run"`
	ChangeFlowRunState *FlowRunStateAction     `tfsdk:"change-flow-run-state"`
	PauseWorkQueue     *WorkQueueAction        `tfsdk:"pause-work-queue"`
	ResumeWorkQueue    *WorkQueueAction        `tfsdk:"resume-work-queue"`
	PauseWorkPool      *WorkPoolAction         `tfsdk:"pause-work-pool"`
	ResumeWorkPool     *WorkPoolAction         `tfsdk:"resume-work-pool"`
	PauseAutomation    *AutomationAction       `tfsdk:"pause-automation"`
	ResumeAutomation   *AutomationAction       `tfsdk:"resume-automation"`
	SuspendFlowRun     *EmptyAction            `tfsdk:"suspend-flow-run"`
	ResumeFlowRun      *EmptyAction            `tfsdk:"resume-flow-run"`
	DeclareIncident    *EmptyAction            `tfsdk:"declare-incident"`
	DoNothing          *EmptyAction            `tfsdk:"do-nothing"`
}

// Common action types
type RunDeploymentAction struct {
	Source       types.String `tfsdk:"source"`
	DeploymentID types.String `tfsdk:"deployment_id"`
	Parameters   types.Map    `tfsdk:"parameters"`
	JobVariables types.Map    `tfsdk:"job_variables"`
}

type SendNotificationAction struct {
	BlockDocumentID types.String `tfsdk:"block_document_id"`
	Subject         types.String `tfsdk:"subject"`
	Body            types.String `tfsdk:"body"`
}

type CallWebhookAction struct {
	BlockDocumentID types.String `tfsdk:"block_document_id"`
	Payload         types.String `tfsdk:"payload"`
}

type FlowRunStateAction struct {
	Name    types.String `tfsdk:"name"`
	State   types.String `tfsdk:"state"`
	Message types.String `tfsdk:"message"`
}

// Actions that require source/id pattern
type DeploymentAction struct {
	Source       types.String `tfsdk:"source"`
	DeploymentID types.String `tfsdk:"deployment_id"`
}

type WorkQueueAction struct {
	Source      types.String `tfsdk:"source"`
	WorkQueueID types.String `tfsdk:"work_queue_id"`
}

type WorkPoolAction struct {
	Source     types.String `tfsdk:"source"`
	WorkPoolID types.String `tfsdk:"work_pool_id"`
}

type AutomationAction struct {
	Source       types.String `tfsdk:"source"`
	AutomationID types.String `tfsdk:"automation_id"`
}

// For actions that don't require any configuration
type EmptyAction struct{}
