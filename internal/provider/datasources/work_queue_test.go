package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
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
}

resource "prefect_work_queue" "test" {
    name = "%s"
	work_pool_name = prefect_work_pool.test.name
	priority = 1
	description = "my work queue"
}

data "prefect_work_queue" "test" {
	name = prefect_work_queue.test.name
	work_pool_name = prefect_work_pool.test.name
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
}

resource "prefect_work_queue" "test_queue1" {
	name = "%s"
	work_pool_name = prefect_work_pool.test_multi.name
	priority = 1
	description = "my work queue"
}

resource "prefect_work_queue" "test_queue2" {
    name = "%s"
	work_pool_name ="%s"
	priority = 2
	description = "my work queue 2"
}

data "prefect_work_queues" "test" {
    work_pool_name = "%s"
}

`, workspace, workPoolName, workQueue1Name, workQueue2Name, workPoolName, workPoolName)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_work_queue(t *testing.T) {
	singleWorkQueueDatasourceName := "data.prefect_work_queue.test"
	multipleWorkQueueDatasourceName := "data.prefect_work_queues.test"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that we can query a single work pool
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
				// Check that we can query multiple work pools
				Config: fixtureAccMultipleWorkQueue(workspace.Resource, "test-pool-multi", "test-queue", "test-queue-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(multipleWorkQueueDatasourceName, "work_queues.#", "3"),
					resource.TestCheckResourceAttr(multipleWorkQueueDatasourceName, "work_queues.0.name", "default1"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.id"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.created"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.created"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.updated"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.is_paused"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.priority"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.0.description"),
					resource.TestCheckResourceAttr(multipleWorkQueueDatasourceName, "work_queues", "test-queue"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.id"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.created"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.updated"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.is_paused"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.priority"),
					resource.TestCheckResourceAttrSet(multipleWorkQueueDatasourceName, "work_queues.1.description"),
				),
			},
		},
	})
}
