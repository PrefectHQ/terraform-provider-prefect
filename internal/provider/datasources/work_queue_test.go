package datasources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
	"k8s.io/utils/ptr"
)

func fixtureAccSingleWorkQueue(
	workspace string,
	workPoolName string,
	workQueueName string,
) string {
	return fmt.Sprintf(`
%s

resource "prefect_work_pool" "test" {
  name = "%s"
  type = "kubernetes"
  paused = "false"
  workspace_id = prefect_workspace.test.id
  depends_on = [prefect_workspace.test]
}

resource "prefect_work_queue" "test" {
  name = "%s"
  work_pool_name = prefect_work_pool.test.name
  priority = 1
  description = "my work queue"
  workspace_id = prefect_workspace.test.id
}

data "prefect_work_queue" "test" {
  name = prefect_work_queue.test.name
  work_pool_name = prefect_work_pool.test.name
  workspace_id = prefect_workspace.test.id
}

`, workspace, workPoolName, workQueueName)
}

func fixtureAccMultipleWorkQueue(
	workspace string,
	workPoolName string,
	workQueue1Name string,
	workQueue2Name string,
) string {
	return fmt.Sprintf(`
%s

resource "prefect_work_pool" "test_multi" {
  name = "%s"
  type = "kubernetes"
  paused = "false"
  workspace_id = prefect_workspace.test.id
  depends_on = [prefect_workspace.test]
}

resource "prefect_work_queue" "test_queue1" {
  name = "%s"
  work_pool_name = prefect_work_pool.test_multi.name
  priority = 1
  description = "my work queue"
  workspace_id = prefect_workspace.test.id
}

resource "prefect_work_queue" "test_queue2" {
  name = "%s"
  work_pool_name = prefect_work_pool.test_multi.name
  workspace_id = prefect_workspace.test.id
}

data "prefect_work_queues" "test" {
  work_pool_name = prefect_work_pool.test_multi.name
  workspace_id = prefect_workspace.test.id
  depends_on = [
    prefect_work_pool.test_multi,
    prefect_work_queue.test_queue1,
    prefect_work_queue.test_queue2
  ]
}

`, workspace, workPoolName, workQueue1Name, workQueue2Name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_work_queue(t *testing.T) {
	singleWorkQueueDatasourceName := "data.prefect_work_queue.test"
	workspace := testutils.NewEphemeralWorkspace()
	workQueues := []*api.WorkQueue{}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that we can query a single work queue
				Config: fixtureAccSingleWorkQueue(workspace.Resource, "test-pool", "test-queue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(singleWorkQueueDatasourceName, "name", "test-queue"),
					resource.TestCheckResourceAttrSet(singleWorkQueueDatasourceName, "id"),
					resource.TestCheckResourceAttrSet(singleWorkQueueDatasourceName, "updated"),
					resource.TestCheckResourceAttr(singleWorkQueueDatasourceName, "is_paused", "false"),
					resource.TestCheckResourceAttr(singleWorkQueueDatasourceName, "priority", "1"),
					resource.TestCheckResourceAttr(singleWorkQueueDatasourceName, "description", "my work queue"),
				),
			},
			{
				// Check that we can query multiple work queues
				Config: fixtureAccMultipleWorkQueue(workspace.Resource, "test-pool-multi", "test-queue", "test-queue-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckworkQueueExists("prefect_work_pool.test_multi", &workQueues),
					testAccCheckWorkQueueValues(&workQueues, expectedWorkQueues),
				),
			},
		},
	})
}

// testAccCheckworkQueueExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckworkQueueExists(workPoolResourceName string, workQueues *[]*api.WorkQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		workPoolResource, exists := s.RootModule().Resources[workPoolResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workPoolResourceName)
		}

		workPoolName := workPoolResource.Primary.Attributes["name"]

		workspaceResource, exists := s.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}

		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, err := testutils.NewTestClient()
		if err != nil {
			return fmt.Errorf("error creating new test client: %w", err)
		}

		workQueuesClient, err := c.WorkQueues(uuid.Nil, workspaceID, workPoolName)
		if err != nil {
			return fmt.Errorf("error creating new work queues client: %w", err)
		}

		fetchedWorkQueues, err := workQueuesClient.List(context.Background(), api.WorkQueueFilter{})
		if err != nil {
			return fmt.Errorf("error fetching workQueues: %w", err)
		}

		if len(fetchedWorkQueues) == 0 {
			return fmt.Errorf("unable to list any work queues for work pool %s", workPoolName)
		}

		*workQueues = append(*workQueues, fetchedWorkQueues...)

		return nil
	}
}

var expectedWorkQueues = []*api.WorkQueue{
	{
		Name:        "default",
		Priority:    ptr.To(int64(2)),
		Description: ptr.To("The work pool's default queue."),
		IsPaused:    false,
	},
	{
		Name:        "test-queue",
		Priority:    ptr.To(int64(1)),
		Description: ptr.To("my work queue"),
		IsPaused:    false,
	},
	{
		Name:        "test-queue-2",
		Description: ptr.To(""),
		IsPaused:    false,
		// Intentionally not setting Priority.
	},
}

// testAccCheckWorkQueueValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckWorkQueueValues(fetched *[]*api.WorkQueue, expected []*api.WorkQueue) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if len(*fetched) != len(expected) {
			return fmt.Errorf("Expected work queues to be %d, got %d", len(expected), len(*fetched))
		}

		for _, expectedWorkQueue := range expected {
			found := false

			for _, fetchedWorkQueue := range *fetched {
				if fetchedWorkQueue.Name != expectedWorkQueue.Name {
					continue
				}

				// Mark the work queue as 'found', and check each of the other fields to ensure they match.
				found = true
				name := expectedWorkQueue.Name

				if *fetchedWorkQueue.Description != *expectedWorkQueue.Description {
					return fmt.Errorf("Expected work queue '%s' description to be %s, got %s", name, *expectedWorkQueue.Description, *fetchedWorkQueue.Description)
				}

				if fetchedWorkQueue.IsPaused != expectedWorkQueue.IsPaused {
					return fmt.Errorf("Expected work queue '%s' is paused to be %t, got %t", name, expectedWorkQueue.IsPaused, fetchedWorkQueue.IsPaused)
				}

				// Priority is special because if one is not configured, it will get a value based on the priority of the most recently created work queue.
				// Because of this, let's only check for matching priority when we configure one.
				if expectedWorkQueue.Priority != nil {
					if *fetchedWorkQueue.Priority != *expectedWorkQueue.Priority {
						return fmt.Errorf("Expected work queue '%s' priority to be %d, got %d", name, *expectedWorkQueue.Priority, *fetchedWorkQueue.Priority)
					}
				}

				break
			}

			if !found {
				return fmt.Errorf("Expected to find work queue '%s' by name", expectedWorkQueue.Name)
			}
		}

		return nil
	}
}
