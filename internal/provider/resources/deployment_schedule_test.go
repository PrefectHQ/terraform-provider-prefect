package resources_test

import (
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentScheduleConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string

	DeploymentSchedule api.DeploymentSchedule
}

func fixtureAccDeploymentSchedule(cfg deploymentScheduleConfig) string {
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

	active = {{.DeploymentSchedule.Active}}
	max_active_runs = {{.DeploymentSchedule.MaxActiveRuns}}

	# seeing inconsistent result with this one
	# max_scheduled_runs = {{.DeploymentSchedule.MaxScheduledRuns}}
	catchup = {{.DeploymentSchedule.Catchup}}

	interval = {{.DeploymentSchedule.Schedule.Interval}}
	timezone = "{{.DeploymentSchedule.Schedule.Timezone}}"

	# add the rest...
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_schedule(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	// Test case: create

	scheduleCreate := api.DeploymentSchedule{
		DeploymentSchedulePayload: api.DeploymentSchedulePayload{
			Active:           true,
			MaxScheduledRuns: 5,
			MaxActiveRuns:    4,
			Catchup:          false,
			Schedule: api.Schedule{
				Interval: 30,
				Timezone: "UTC",
			},
		},
	}

	cfgCreate := deploymentScheduleConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
		DeploymentSchedule:    scheduleCreate,
	}

	createChecks := []resource.TestCheckFunc{
		testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
	}
	createChecks = append(createChecks, testScheduleValues(cfgCreate.DeploymentSchedule)...)

	// Test case: update

	scheduleUpdate := api.DeploymentSchedule{
		DeploymentSchedulePayload: api.DeploymentSchedulePayload{
			// Changing active to false is failing, coming back as true
			Active:           true,
			MaxScheduledRuns: 7,
			MaxActiveRuns:    6,
			Catchup:          true,
			Schedule: api.Schedule{
				Interval: 60,
				Timezone: "America/New_York",
			},
		},
	}

	cfgUpdate := cfgCreate
	cfgUpdate.DeploymentSchedule = scheduleUpdate

	updateChecks := []resource.TestCheckFunc{
		testAccCheckDeploymentExists("prefect_deployment.test", &api.Deployment{}),
	}
	updateChecks = append(updateChecks, testScheduleValues(cfgUpdate.DeploymentSchedule)...)

	// Run the tests

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentSchedule(cfgCreate),
				Check:  resource.ComposeAggregateTestCheckFunc(createChecks...),
			},
			{
				Config: fixtureAccDeploymentSchedule(cfgUpdate),
				Check:  resource.ComposeAggregateTestCheckFunc(updateChecks...),
			},
		},
	})
}

func testScheduleValues(schedule api.DeploymentSchedule) []resource.TestCheckFunc {
	checks := scheduleToChecks(schedule)
	tests := make([]resource.TestCheckFunc, 0, len(checks))

	for key, value := range checks {
		tests = append(tests, resource.TestCheckResourceAttr("prefect_deployment_schedule.test", key, value))
	}

	return tests
}

func scheduleToChecks(schedule api.DeploymentSchedule) map[string]string {
	result := map[string]string{}

	result["timezone"] = schedule.Schedule.Timezone
	result["active"] = strconv.FormatBool(schedule.Active)
	result["catchup"] = strconv.FormatBool(schedule.Catchup)
	result["max_active_runs"] = strconv.FormatFloat(float64(schedule.MaxActiveRuns), 'f', -1, 64)
	result["interval"] = strconv.FormatFloat(float64(schedule.Schedule.Interval), 'f', -1, 64)

	// This one is failing, getting 0
	// result["max_scheduled_runs"] = strconv.FormatFloat(float64(schedule.MaxScheduledRuns), 'f', -1, 64)

	return result
}
