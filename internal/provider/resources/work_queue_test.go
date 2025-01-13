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

func fixtureAccWorkQueueCreate(workspace, name, poolType, baseJobTemplate string, paused bool) string {
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
	priority = 1
	description = "work queue"
}

`, workspace, name, name, poolType, paused, baseJobTemplate, name, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_queue(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	randomName := testutils.NewRandomPrefixedString()

	workQueueResourceName := "prefect_work_queue." + randomName
	workQueueDescription := "work queue"
	workQueuePriority := int64(1)
	poolType := "kubernetes"

	baseJobTemplate := fmt.Sprintf(baseJobTemplateTpl, "The name given to infrastructure created by a worker.")

	var workQueue api.WorkQueue

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the work queue resource
				Config: fixtureAccWorkQueueCreate(workspace.Resource, randomName, poolType, baseJobTemplate, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkQueueExists(workQueueResourceName, &workQueue, randomName),
					testAccCheckWorkQueueValues(
						&workQueue,
						&api.WorkQueue{
							Name:        randomName,
							IsPaused:    true,
							Priority:    &workQueuePriority,
							Description: &workQueueDescription,
						},
					),
					resource.TestCheckResourceAttr(workQueueResourceName, "name", randomName),
					resource.TestCheckResourceAttr(workQueueResourceName, "paused", "true"),
				),
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

		if fetchedWorkQueue.Priority != valuesToCheck.Priority {
			return fmt.Errorf("Expected work queue type to be %d, got %d", valuesToCheck.Priority, fetchedWorkQueue.Priority)
		}

		if fetchedWorkQueue.Description != valuesToCheck.Description {
			return fmt.Errorf("Expected work queue type to be %s, got %s", *valuesToCheck.Description, *fetchedWorkQueue.Description)
		}

		if fetchedWorkQueue.IsPaused != valuesToCheck.IsPaused {
			return fmt.Errorf("Expected work queue paused to be %t, got %t", valuesToCheck.IsPaused, fetchedWorkQueue.IsPaused)
		}

		return nil
	}
}
