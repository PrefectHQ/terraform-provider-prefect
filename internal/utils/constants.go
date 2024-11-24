package utils

// Workspace Accessor Types.
const ServiceAccount string = "SERVICE_ACCOUNT"
const User string = "USER"
const Team string = "TEAM"

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

var AllMetricNames = []string{
	"lateness",
	"duration",
	"successes",
}
