package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentScheduleConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string

	api.DeploymentSchedule
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

	active = {{.Active}}
	max_active_runs = {{.MaxActiveRuns}}

	# seeing inconsistent result with this one
	# max_scheduled_runs = {{.MaxScheduledRuns}}
	catchup = {{.Catchup}}

	interval = {{.Schedule.Interval}}
	timezone = "{{.Schedule.Timezone}}"

	# add the rest...
}
`

	result := helpers.RenderTemplate(tmpl, cfg)
	fmt.Println(result)
	return result
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_schedule(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	schedule := api.DeploymentSchedule{
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
		DeploymentSchedule:    schedule,
	}

	var deployment api.Deployment
	var deploymentSchedules api.DeploymentSchedule

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentSchedule(cfgCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", &deployment),
					testAccCheckDeploymentScheduleExists("prefect_deployment_schedule.test", &deploymentSchedules),
					// testAccCheckDeploymentScheduleValues([]*api.DeploymentSchedule{&deploymentSchedules}, []*api.DeploymentSchedule{&schedule}),
				),
			},
		},
	})
}

// testAccCheckDeploymentScheduleExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckDeploymentScheduleExists(deploymentScheduleResourceName string, deploymentSchedule *api.DeploymentSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the deployment schedule resource we just created from the state
		deploymentScheduleResource, exists := s.RootModule().Resources[deploymentScheduleResourceName]
		if !exists {
			return fmt.Errorf("deployment schedule resource not found: %s", deploymentScheduleResourceName)
		}
		deploymentScheduleID, _ := uuid.Parse(deploymentScheduleResource.Primary.Attributes["deployment_id"])

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("workspace resource not found: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		deploymentScheduleClient, _ := c.DeploymentSchedule(uuid.Nil, workspaceID)

		fetchedDeploymentSchedules, err := deploymentScheduleClient.Read(context.Background(), deploymentScheduleID)
		if err != nil {
			return fmt.Errorf("error fetching deployment schedules: %w", err)
		}

		// fetchedSchedule, err := scheduleFound(fetchedDeploymentSchedules, []*api.DeploymentSchedule{deploymentSchedule})
		if len(fetchedDeploymentSchedules) == 0 {
			return fmt.Errorf("deployment schedule %s not found", deploymentScheduleID)
		}
		fetchedSchedule := fetchedDeploymentSchedules[0]

		// Assign the fetched deployment schedule to the passed pointer
		// so we can use it in the next test assertion
		*deploymentSchedule = *fetchedSchedule

		return nil
	}
}

// testAccCheckDeploymentValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckDeploymentScheduleValues(fetchedDeploymentSchedules, expectedDeploymentSchedules []*api.DeploymentSchedule) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		_, err := scheduleFound(fetchedDeploymentSchedules, expectedDeploymentSchedules)
		if err != nil {
			return fmt.Errorf("expected deployment schedule %s not found: %w", expectedDeploymentSchedules[0].ID, err)
		}

		return nil
	}
}

func scheduleFound(fetched []*api.DeploymentSchedule, expected []*api.DeploymentSchedule) (*api.DeploymentSchedule, error) {
	if len(fetched) != len(expected) {
		return nil, fmt.Errorf("got %d schedules, expected %d", len(fetched), len(expected))
	}

	var result *api.DeploymentSchedule

	for i := range expected {
		found := false
		for j := range fetched {
			if fetched[j].ID == expected[i].ID {
				found = true
				result = fetched[j]

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("schedule %s not found", expected[i].ID)
		}
	}

	return result, nil
}
