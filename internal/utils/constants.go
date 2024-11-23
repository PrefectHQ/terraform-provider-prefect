package utils

// Workspace Accessor Types.
const ServiceAccount string = "SERVICE_ACCOUNT"
const User string = "USER"
const Team string = "TEAM"

// Automation Action Types.

const (
	ActionDoNothing          string = "do-nothing"
	ActionRunDeployment      string = "run-deployment"
	ActionPauseDeployment    string = "pause-deployment"
	ActionResumeDeployment   string = "resume-deployment"
	ActionCancelFlowRun      string = "cancel-flow-run"
	ActionChangeFlowRunState string = "change-flow-run-state"
	ActionPauseWorkQueue     string = "pause-work-queue"
	ActionResumeWorkQueue    string = "resume-work-queue"
	ActionSendNotification   string = "send-notification"
	ActionCallWebhook        string = "call-webhook"
	ActionPauseAutomation    string = "pause-automation"
	ActionResumeAutomation   string = "resume-automation"
	ActionSuspendFlowRun     string = "suspend-flow-run"
	ActionResumeFlowRun      string = "resume-flow-run"
	ActionDeclareIncident    string = "declare-incident"
	ActionPauseWorkPool      string = "pause-work-pool"
	ActionResumeWorkPool     string = "resume-work-pool"
)

var AllAutomationActionTypes = []string{
	ActionDoNothing,
	ActionRunDeployment,
	ActionPauseDeployment,
	ActionResumeDeployment,
	ActionCancelFlowRun,
	ActionChangeFlowRunState,
	ActionPauseWorkQueue,
	ActionResumeWorkQueue,
	ActionSendNotification,
	ActionCallWebhook,
	ActionPauseAutomation,
	ActionResumeAutomation,
	ActionSuspendFlowRun,
	ActionResumeFlowRun,
	ActionDeclareIncident,
	ActionPauseWorkPool,
	ActionResumeWorkPool,
}
