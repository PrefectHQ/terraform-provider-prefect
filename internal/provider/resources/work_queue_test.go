package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkQueueCreate(
	workspace string,
	workPoolName string,
	poolType string,
	baseJobTemplate string,
	paused bool,
	workQueueName string,
	priority int64,
	description string,
) string {
	return fmt.Sprintf(`
%s

resource "prefect_work_pool" "%s" {
  name = "%s"
  type = "%s"
  paused = %t
  base_job_template = jsonencode(%s)
  workspace_id = prefect_workspace.test.id
  depends_on = [prefect_workspace.test]
}

resource "prefect_work_queue" "%s" {
  name = "%s"
  work_pool_name = prefect_work_pool.%s.name
  priority = %d
  description = "%s"
  workspace_id = prefect_workspace.test.id
}

`, workspace, workPoolName, workPoolName, poolType, paused, baseJobTemplate, workQueueName, workQueueName, workPoolName, priority, description)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_queue(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	workPoolName := testutils.NewRandomPrefixedString()

	workQueueName := testutils.NewRandomPrefixedString()
	workQueueName2 := testutils.NewRandomPrefixedString()

	workQueueResourceName := "prefect_work_queue." + workQueueName
	workQueueResourceName2 := "prefect_work_queue." + workQueueName2

	workQueueDescriptionFirst := "work queue"
	workQueueDescriptionSecond := "work queue updated"

	workQueuePriorityFirst := int64(1)
	workQueuePrioritySecond := int64(2)

	poolType := "kubernetes"

	baseJobTemplate := fmt.Sprintf(baseJobTemplateTplForQueues, "The name given to infrastructure created by a worker.")

	var workQueue api.WorkQueue

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the work queue resource
				Config: fixtureAccWorkQueueCreate(
					workspace.Resource,
					workPoolName,
					poolType,
					baseJobTemplate,
					true,
					workQueueName,
					workQueuePriorityFirst,
					workQueueDescriptionFirst,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkQueueExists(workQueueResourceName, &workQueue, workPoolName),
					testAccCheckWorkQueueValues(
						&workQueue,
						&api.WorkQueue{
							Name:        workQueueName,
							IsPaused:    false,
							Priority:    &workQueuePriorityFirst,
							Description: &workQueueDescriptionFirst,
						},
					),
					resource.TestCheckResourceAttr(workQueueResourceName, "name", workQueueName),
					resource.TestCheckResourceAttr(workQueueResourceName, "priority", "1"),
					resource.TestCheckResourceAttr(workQueueResourceName, "description", workQueueDescriptionFirst),
					resource.TestCheckResourceAttr(workQueueResourceName, "is_paused", "false"),
				),
			},
			{
				// Check that changing the priority will update the resource in place
				Config: fixtureAccWorkQueueCreate(
					workspace.Resource,
					workPoolName,
					poolType,
					baseJobTemplate,
					true,
					workQueueName,
					workQueuePrioritySecond,
					workQueueDescriptionFirst,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueIDAreEqual(workQueueResourceName, &workQueue),
					testAccCheckWorkQueueExists(workQueueResourceName, &workQueue, workPoolName),
					testAccCheckWorkQueueValues(
						&workQueue,
						&api.WorkQueue{
							Name:        workQueueName,
							IsPaused:    false,
							Priority:    &workQueuePrioritySecond,
							Description: &workQueueDescriptionFirst,
						},
					),
					resource.TestCheckResourceAttr(workQueueResourceName, "name", workQueueName),
					resource.TestCheckResourceAttr(workQueueResourceName, "priority", "2"),
					resource.TestCheckResourceAttr(workQueueResourceName, "description", workQueueDescriptionFirst),
					resource.TestCheckResourceAttr(workQueueResourceName, "is_paused", "false"),
				),
			},
			{
				// Check that changing the Description will update the resource in place
				Config: fixtureAccWorkQueueCreate(
					workspace.Resource,
					workPoolName,
					poolType,
					baseJobTemplate,
					true,
					workQueueName,
					workQueuePrioritySecond,
					workQueueDescriptionSecond,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueIDAreEqual(workQueueResourceName, &workQueue),
					testAccCheckWorkQueueExists(workQueueResourceName, &workQueue, workPoolName),
					testAccCheckWorkQueueValues(
						&workQueue,
						&api.WorkQueue{
							Name:        workQueueName,
							IsPaused:    false,
							Priority:    &workQueuePrioritySecond,
							Description: &workQueueDescriptionSecond,
						},
					),
					resource.TestCheckResourceAttr(workQueueResourceName, "name", workQueueName),
					resource.TestCheckResourceAttr(workQueueResourceName, "priority", "2"),
					resource.TestCheckResourceAttr(workQueueResourceName, "description", workQueueDescriptionSecond),
					resource.TestCheckResourceAttr(workQueueResourceName, "is_paused", "false"),
				),
			},
			{
				// Check that changing the name will re-create the resource
				Config: fixtureAccWorkQueueCreate(
					workspace.Resource,
					workPoolName,
					poolType,
					baseJobTemplate,
					true,
					workQueueName2,
					workQueuePrioritySecond,
					workQueueDescriptionSecond,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueIDsNotEqual(workQueueResourceName2, &workQueue),
					testAccCheckWorkQueueExists(workQueueResourceName2, &workQueue, workPoolName),
					testAccCheckWorkQueueValues(
						&workQueue,
						&api.WorkQueue{
							Name:        workQueueName2,
							IsPaused:    false,
							Priority:    &workQueuePrioritySecond,
							Description: &workQueueDescriptionSecond,
						},
					),
				),
			},
			// Import State checks - import by workspace_id,name, work_pool_name (dynamic)
			{
				ImportState:       true,
				ResourceName:      workQueueResourceName2,
				ImportStateIdFunc: getWorkQueueImportStateID(workQueueResourceName2, workPoolName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWorkQueueExists(workQueueResourceName string, workQueue *api.WorkQueue, workPoolName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workQueueResource, exists := state.RootModule().Resources[workQueueResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workQueueResourceName)
		}

		workspaceResource, exists := state.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		WorkQueuesClient, _ := c.WorkQueues(uuid.Nil, workspaceID, workPoolName)

		workQueueName := workQueueResource.Primary.Attributes["name"]

		fetchedWorkQueue, err := WorkQueuesClient.Get(context.Background(), workQueueName)
		if err != nil {
			return fmt.Errorf("Error fetching work queue: %w", err)
		}
		if fetchedWorkQueue == nil {
			return fmt.Errorf("Work Queue not found for name: %s", workQueueName)
		}

		*workQueue = *fetchedWorkQueue

		return nil
	}
}

func testAccCheckWorkQueueValues(fetchedWorkQueue *api.WorkQueue, valuesToCheck *api.WorkQueue) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedWorkQueue.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected work queue name to be %s, got %s", valuesToCheck.Name, fetchedWorkQueue.Name)
		}

		if *fetchedWorkQueue.Priority != *valuesToCheck.Priority {
			return fmt.Errorf("Expected work queue Priority to be %d, got %d", *valuesToCheck.Priority, *fetchedWorkQueue.Priority)
		}

		if *fetchedWorkQueue.Description != *valuesToCheck.Description {
			return fmt.Errorf("Expected work queue Description to be %s, got %s", *valuesToCheck.Description, *fetchedWorkQueue.Description)
		}

		if fetchedWorkQueue.IsPaused != valuesToCheck.IsPaused {
			return fmt.Errorf("Expected work queue paused to be %t, got %t", valuesToCheck.IsPaused, fetchedWorkQueue.IsPaused)
		}

		return nil
	}
}

func testAccCheckQueueIDAreEqual(resourceName string, fetchedWorkQueue *api.WorkQueue) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workQueueResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkQueue.ID.String()

		if workQueueResource.Primary.ID != id {
			return fmt.Errorf("Expected %s and %s to be equal", workQueueResource.Primary.ID, id)
		}

		return nil
	}
}

func testAccCheckQueueIDsNotEqual(resourceName string, fetchedWorkQueue *api.WorkQueue) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workQueueResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkQueue.ID.String()

		if workQueueResource.Primary.ID == id {
			return fmt.Errorf("Expected %s and %s to be different", workQueueResource.Primary.ID, id)
		}

		return nil
	}
}

func getWorkQueueImportStateID(workQueueResourceName string, workPoolName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceResource, exists := state.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		workQueueResource, exists := state.RootModule().Resources[workQueueResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workQueueResourceName)
		}
		workQueueName := workQueueResource.Primary.Attributes["name"]

		return fmt.Sprintf("%s, %s,%s", workspaceID, workPoolName, workQueueName), nil
	}
}

var baseJobTemplateTplForQueues = `
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
