package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type fixtureConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
}

func fixtureAccDeploymentScheduleInterval(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	active = true
	catchup = false
	max_active_runs = 10
	timezone = "America/New_York"

	interval = 30
	anchor_date = "2024-01-01T00:00:00Z"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleIntervalUpdate(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	# Update these values
	active = false
	catchup = true
	max_active_runs = 20
	timezone = "America/Chicago"

	interval = 30
	anchor_date = "2024-01-01T00:00:00Z"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleCron(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	cron = "* * * * *"
	day_or = true
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleRRule(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	rrule = "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccDeploymentScheduleMultiple(cfg fixtureConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test_cron" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	cron = "* * * * *"
	day_or = true
}

resource "prefect_deployment_schedule" "test_rrule" {
	workspace_id = prefect_workspace.test.id
	deployment_id = prefect_deployment.test.id

	rrule = "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"
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
	}

	resourceName := "prefect_deployment_schedule.test"

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
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueBool(resourceName, "active", true),
					testutils.ExpectKnownValueBool(resourceName, "catchup", false),
					testutils.ExpectKnownValueNumber(resourceName, "interval", 30),
					testutils.ExpectKnownValue(resourceName, "timezone", "America/New_York"),
				},
			},
			{
				// Test interval schedule update
				Config: fixtureAccDeploymentScheduleIntervalUpdate(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueBool(resourceName, "active", false),
					testutils.ExpectKnownValueBool(resourceName, "catchup", true),
					testutils.ExpectKnownValueNumber(resourceName, "max_active_runs", 20),
					testutils.ExpectKnownValue(resourceName, "timezone", "America/Chicago"),
				},
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
					// RRule schedule tests
					testutils.ExpectKnownValue("prefect_deployment_schedule.test_rrule", "rrule", "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"),
				},
			},
		},
	})
}
