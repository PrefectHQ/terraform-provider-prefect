//nolint:revive // we can rename this in a future change
package utils

// Workspace Accessor Types.
const ServiceAccount string = "SERVICE_ACCOUNT"
const User string = "USER"
const Team string = "TEAM"

// Automation Trigger Types.
const TriggerTypeMetric string = "metric"
const TriggerTypeEvent string = "event"
const TriggerTypeCompound string = "compound"
const TriggerTypeSequence string = "sequence"

// Automation Action Types.
// Source: https://docs.prefect.io/v3/api-ref/rest-api/server/automations/create-automation#body-actions
var AllAutomationActionTypes = []string{
	"do-nothing",
	"run-deployment",
	"pause-deployment",
	"resume-deployment",
	"cancel-flow-run",
	"change-flow-run-state",
	"pause-work-queue",
	"resume-work-queue",
	"send-notification",
	"send-email-notification",
	"call-webhook",
	"pause-automation",
	"resume-automation",
	"suspend-flow-run",
	"resume-flow-run",
	"declare-incident",
	"pause-work-pool",
	"resume-work-pool",
	"pause-schedule-for-flow-run",
	"resume-schedule-for-flow-run",
}

// CloudOnlyAutomationActionTypes lists action types that are only available on Prefect Cloud.
var CloudOnlyAutomationActionTypes = []string{
	"send-email-notification",
	"pause-schedule-for-flow-run",
	"resume-schedule-for-flow-run",
}

var AllMetricOperators = []string{
	"<",
	">",
	">=",
	"<=",
}

var AllTriggerMetricNames = []string{
	"lateness",
	"duration",
	"successes",
}
