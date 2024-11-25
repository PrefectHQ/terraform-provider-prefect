package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
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

	return helpers.RenderTemplate(tmpl, cfg)
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

	return helpers.RenderTemplate(tmpl, cfg)
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

	return helpers.RenderTemplate(tmpl, cfg)
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

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_schedule(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	fixtureCfg := fixtureConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			// Test interval schedule
			{
				Config: fixtureAccDeploymentScheduleInterval(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "active", "true"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "catchup", "false"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "interval", "30"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "timezone", "America/New_York"),
				),
			},
			// Test interval schedule update
			{
				Config: fixtureAccDeploymentScheduleIntervalUpdate(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "active", "false"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "catchup", "true"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "max_active_runs", "20"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "timezone", "America/Chicago"),
				),
			},
			// Test cron schedule
			{
				Config: fixtureAccDeploymentScheduleCron(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "cron", "* * * * *"),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "day_or", "true"),
				),
			},
			// Test rrule schedule
			{
				Config: fixtureAccDeploymentScheduleRRule(fixtureCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
					resource.TestCheckResourceAttr("prefect_deployment_schedule.test", "rrule", "FREQ=DAILY;BYHOUR=10;BYMINUTE=30"),
				),
			},
		},
	})
}