package resources_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

const workPoolWithoutWorkspaceID = `
resource "prefect_work_pool" "invalid_work_pool" {
	name = "invalid-work-pool"
	type = "kubernetes"
}
`

func fixtureAccWorkPoolCreate(workspace, workspaceIDArg, name, poolType, baseJobTemplate string, paused bool) string {
	return fmt.Sprintf(`
%[1]s

resource "prefect_work_pool" "%[3]s" {
	name = "%[3]s"
	type = "%[4]s"
	paused = %[6]t
	base_job_template = jsonencode(%[5]s)
	%[2]s
}
`, workspace, workspaceIDArg, name, poolType, baseJobTemplate, paused)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_pool(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	randomName := testutils.NewRandomPrefixedString()
	workPoolResourceName := "prefect_work_pool." + randomName

	randomName2 := testutils.NewRandomPrefixedString()
	workPoolResourceName2 := "prefect_work_pool." + randomName2

	poolType := "kubernetes"
	poolType2 := "ecs"

	baseJobTemplate := fmt.Sprintf(baseJobTemplateTpl, "The name given to infrastructure created by a worker.")
	baseJobTemplateExpected := testutils.NormalizedValueForJSON(t, baseJobTemplate)

	baseJobTemplate2 := fmt.Sprintf(baseJobTemplateTpl, "The name given to infrastructure created by a worker!")
	baseJobTemplateExpected2 := testutils.NormalizedValueForJSON(t, baseJobTemplate2)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workPool api.WorkPool

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that workspace_id missing causes a failure
				Config:      workPoolWithoutWorkspaceID,
				ExpectError: regexp.MustCompile(".*require an account_id and workspace_id to be set.*"),
				SkipFunc:    testutils.SkipFuncOSS,
			},
			{
				// Check creation + existence of the work pool resource
				Config: fixtureAccWorkPoolCreate(workspace.Resource, workspace.IDArg, randomName, poolType, baseJobTemplate, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkPoolExists(workPoolResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName, Type: poolType, IsPaused: true}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(workPoolResourceName, "base_job_template", baseJobTemplateExpected),
					testutils.ExpectKnownValue(workPoolResourceName, "name", randomName),
					testutils.ExpectKnownValue(workPoolResourceName, "type", poolType),
					testutils.ExpectKnownValueBool(workPoolResourceName, "paused", true),
				},
			},
			{
				// Check that changing the paused state will update the resource in place
				Config: fixtureAccWorkPoolCreate(workspace.Resource, workspace.IDArg, randomName, poolType, baseJobTemplate, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDAreEqual(workPoolResourceName, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName, Type: poolType, IsPaused: false}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(workPoolResourceName, "name", randomName),
					testutils.ExpectKnownValue(workPoolResourceName, "type", poolType),
					testutils.ExpectKnownValueBool(workPoolResourceName, "paused", false),
				},
			},
			{
				// Check that changing the baseJobTemplate will update the resource in place
				Config: fixtureAccWorkPoolCreate(workspace.Resource, workspace.IDArg, randomName, poolType, baseJobTemplate2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDAreEqual(workPoolResourceName, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName, &workPool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(workPoolResourceName, "base_job_template", baseJobTemplateExpected2),
				},
			},
			{
				// Check that changing the name will re-create the resource
				Config: fixtureAccWorkPoolCreate(workspace.Resource, workspace.IDArg, randomName2, poolType, baseJobTemplate2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDsNotEqual(workPoolResourceName2, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName2, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName2, Type: poolType, IsPaused: false}),
				),
			},
			{
				// Check that changing the poolType will re-create the resource
				Config: fixtureAccWorkPoolCreate(workspace.Resource, workspace.IDArg, randomName2, poolType2, baseJobTemplate2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDsNotEqual(workPoolResourceName2, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName2, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName2, Type: poolType2, IsPaused: false}),
				),
			},
			// Import State checks - import by workspace_id,name (dynamic)
			{
				ImportState:             true,
				ResourceName:            workPoolResourceName2,
				ImportStateIdFunc:       testutils.GetResourceWorkspaceImportStateIDByName(workPoolResourceName2),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"base_job_template"}, // we've already tested this, and we can't provide our unique equality check here
			},
		},
	})
}

func testAccCheckWorkPoolExists(workPoolResourceName string, workPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolName, err := testutils.GetResourceAttributeFromStateByAttribute(state, workPoolResourceName, "name")
		if err != nil {
			return fmt.Errorf("error fetching work pool name: %w", err)
		}

		var workspaceID uuid.UUID

		if !testutils.TestContextOSS() {
			// Get the workspace resource we just created from the state
			workspaceID, err = testutils.GetResourceWorkspaceIDFromState(state)
			if err != nil {
				return fmt.Errorf("error fetching workspace ID: %w", err)
			}
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		workPoolsClient, _ := c.WorkPools(uuid.Nil, workspaceID)

		fetchedWorkPool, err := workPoolsClient.Get(context.Background(), workPoolName)
		if err != nil {
			return fmt.Errorf("Error fetching work pool: %w", err)
		}
		if fetchedWorkPool == nil {
			return fmt.Errorf("Work Pool not found for name: %s", workPoolName)
		}

		*workPool = *fetchedWorkPool

		return nil
	}
}

func testAccCheckWorkPoolValues(fetchedWorkPool *api.WorkPool, valuesToCheck *api.WorkPool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedWorkPool.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected work pool name to be %s, got %s", valuesToCheck.Name, fetchedWorkPool.Name)
		}

		if fetchedWorkPool.Type != valuesToCheck.Type {
			return fmt.Errorf("Expected work pool type to be %s, got %s", valuesToCheck.Type, fetchedWorkPool.Type)
		}

		if fetchedWorkPool.IsPaused != valuesToCheck.IsPaused {
			return fmt.Errorf("Expected work pool paused to be %t, got %t", valuesToCheck.IsPaused, fetchedWorkPool.IsPaused)
		}

		return nil
	}
}

func testAccCheckIDAreEqual(resourceName string, fetchedWorkPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkPool.ID.String()

		if workPoolResource.Primary.ID != id {
			return fmt.Errorf("Expected %s and %s to be equal", workPoolResource.Primary.ID, id)
		}

		return nil
	}
}

func testAccCheckIDsNotEqual(resourceName string, fetchedWorkPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkPool.ID.String()

		if workPoolResource.Primary.ID == id {
			return fmt.Errorf("Expected %s and %s to be different", workPoolResource.Primary.ID, id)
		}

		return nil
	}
}

var baseJobTemplateTpl = `
{
  "job_configuration": {
    "command": "{{ command }}",
    "env": "{{ env }}",
    "labels": "{{ labels }}",
    "name": "{{ name }}",
    "stream_output": "{{ stream_output }}",
    "working_dir": "{{ working_dir }}"
  },
  "variables": {
    "type": "object",
    "properties": {
      "name": {
        "title": "Name",
        "description": "%s",
        "type": "string"
      },
      "env": {
        "title": "Environment Variables",
        "default": {
          "PREFECT_CLOUD_API_URL": "http://localhost:8000/"
        },
        "description": "Environment variables to set when starting a flow run.",
        "type": "object",
        "additionalProperties": {
          "type": "string"
        }
      },
      "labels": {
        "title": "Labels",
        "description": "Labels applied to infrastructure created by a worker.",
        "type": "object",
        "additionalProperties": {
          "type": "string"
        }
      },
      "command": {
        "title": "Command",
        "description": "The command to use when starting a flow run. In most cases, this should be left blank and the command will be automatically generated by the worker.",
        "type": "string"
      },
      "stream_output": {
        "title": "Stream Output",
        "description": "If enabled, workers will stream output from flow run processes to local standard output.",
        "default": true,
        "type": "boolean"
      },
      "working_dir": {
        "title": "Working Directory",
        "description": "If provided, workers will open flow run processes within the specified path as the working directory. Otherwise, a temporary directory will be created.",
        "type": "string",
        "format": "path"
      }
    }
  }
}
`
