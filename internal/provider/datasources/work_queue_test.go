package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkQueue() string {
	return `
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
data "prefect_work_pool" "evergreen" {
	name = "evergreen-pool"
	workspace_id = data.prefect_workspace.evergreen.id
}
data "prefect_work_queue" "evergreen" {
	name = "evergreen-queue"
	workspace_id = data.prefect_workspace.evergreen.id
	work_pool_name = data.prefect_work_pool.evergreen.name
}
`
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_work_queue(t *testing.T) {
	WorkQueueDatasourceName := "data.prefect_work_queue.evergreen"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that we can query a work queue
				Config: fixtureAccWorkQueue(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(WorkQueueDatasourceName, "name", "evergreen-queue"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "id"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "created"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "updated"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "is_paused"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "priority"),
					resource.TestCheckResourceAttrSet(WorkQueueDatasourceName, "concurrency_limit"),
				),
			},
		},
	})
}
