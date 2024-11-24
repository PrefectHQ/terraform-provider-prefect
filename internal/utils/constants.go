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
	"call-webhook",
	"pause-automation",
	"resume-automation",
	"suspend-flow-run",
	"resume-flow-run",
	"declare-incident",
	"pause-work-pool",
	"resume-work-pool",
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
