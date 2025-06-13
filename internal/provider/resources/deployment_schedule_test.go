package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

const (
	resourceName = "prefect_deployment_schedule.test"
)

type fixtureConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
	WorkspaceIDArg        string
}

func fixtureAccDeploymentScheduleInterval(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	active = true
	timezone = "America/New_York"

	interval = 30
	anchor_date = "2024-01-01T00:00:00Z"

	parameters = jsonencode({
	 	env = "test"
	 	version = "1.0"
	})
	slug = "test-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleIntervalUpdate(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	# Update these values
	active = false
	timezone = "America/Chicago"

	interval = 30
	anchor_date = "2024-01-01T00:00:00Z"

	parameters = jsonencode({
	 	env = "staging"
	 	version = "2.0"
	})
	slug = "updated-test-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleCron(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	cron = "* * * * *"
	day_or = true

	parameters = jsonencode({
	 	mode = "cron"
	})
	slug = "cron-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleRRule(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	rrule = "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"

	parameters = jsonencode({
	 	mode = "rrule"
	 	repeat = "daily"
	})
	slug = "rrule-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleMultiple(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test_cron" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	cron = "* * * * *"
	day_or = true

	parameters = jsonencode({
	 	schedule_type = "cron"
	})
	slug = "multi-cron-schedule"
}

resource "prefect_deployment_schedule" "test_rrule" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	rrule = "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"

	parameters = jsonencode({
	 	schedule_type = "rrule"
	})
	slug = "multi-rrule-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleWithParameters(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	{{.WorkspaceIDArg}}
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	{{.WorkspaceIDArg}}
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	{{.WorkspaceIDArg}}
	deployment_id = prefect_deployment.test.id

	active = true
	timezone = "UTC"

	interval = 60
	anchor_date = "2024-01-01T00:00:00Z"

	parameters = jsonencode({
		env = "production"
		version = "3.0"
	})
	slug = "parameters-test-schedule"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_schedule(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	fixtureCfg := fixtureConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
		WorkspaceIDArg:        workspace.IDArg,
	}

	fixtureCfgUpdate := fixtureCfg

	stateChecksForInterval := []statecheck.StateCheck{
		testutils.ExpectKnownValueBool(resourceName, "active", true),
		testutils.ExpectKnownValueNumber(resourceName, "interval", 30),
		testutils.ExpectKnownValue(resourceName, "timezone", "America/New_York"),
		testutils.ExpectKnownValue(resourceName, "slug", "test-schedule"),
	}

	stateChecksForIntervalUpdate := []statecheck.StateCheck{
		testutils.ExpectKnownValueBool(resourceName, "active", false),
		testutils.ExpectKnownValue(resourceName, "timezone", "America/Chicago"),
		testutils.ExpectKnownValue(resourceName, "slug", "updated-test-schedule"),
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test interval schedule
				Config: fixtureAccDeploymentScheduleInterval(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: stateChecksForInterval,
			},
			{
				// Test interval schedule update
				Config: fixtureAccDeploymentScheduleIntervalUpdate(fixtureCfgUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: stateChecksForIntervalUpdate,
			},
			{
				// Test cron schedule
				Config: fixtureAccDeploymentScheduleCron(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "cron", "* * * * *"),
					testutils.ExpectKnownValueBool(resourceName, "day_or", true),
					testutils.ExpectKnownValue(resourceName, "slug", "cron-schedule"),
				},
			},
			{
				// Test rrule schedule
				Config: fixtureAccDeploymentScheduleRRule(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "rrule", "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"),
					testutils.ExpectKnownValue(resourceName, "slug", "rrule-schedule"),
				},
			},
			{
				// Test multiple schedules for one deployment
				Config: fixtureAccDeploymentScheduleMultiple(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Cron schedule tests
					testutils.ExpectKnownValue("prefect_deployment_schedule.test_cron", "cron", "* * * * *"),
					testutils.ExpectKnownValueBool("prefect_deployment_schedule.test_cron", "day_or", true),
					testutils.ExpectKnownValue("prefect_deployment_schedule.test_cron", "slug", "multi-cron-schedule"),
					// RRule schedule tests
					testutils.ExpectKnownValue("prefect_deployment_schedule.test_rrule", "rrule", "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"),
					testutils.ExpectKnownValue("prefect_deployment_schedule.test_rrule", "slug", "multi-rrule-schedule"),
				},
			},
			{
				// Test parameters and slug fields
				Config: fixtureAccDeploymentScheduleWithParameters(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "slug", "parameters-test-schedule"),
					testutils.ExpectKnownValueNumber(resourceName, "interval", 60),
					testutils.ExpectKnownValue(resourceName, "timezone", "UTC"),
				},
			},
		},
	})
}
