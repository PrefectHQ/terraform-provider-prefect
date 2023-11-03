package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccSingleWorkPool() string {
	return `
data "prefect_workspace" "evergreen" {
	handle = "evergreen-workspace"
}
data "prefect_work_pool" "evergreen" {
	name = "evergreen-pool"
	workspace_id = data.prefect_workspace.evergreen.id
}
`
}
func fixtureAccMultipleWorkPools() string {
	return `
data "prefect_workspace" "evergreen" {
	handle = "evergreen-workspace"
}
data "prefect_work_pools" "evergreen" {
	workspace_id = data.prefect_workspace.evergreen.id
}
`
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_work_pool(t *testing.T) {
	singleWorkPoolDatasourceName := "data.prefect_work_pool.evergreen"
	multipleWorkPoolDatasourceName := "data.prefect_work_pools.evergreen"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that we can query a single work pool
				Config: fixtureAccSingleWorkPool(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(singleWorkPoolDatasourceName, "name", "evergreen-pool"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "id"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "created"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "updated"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "type"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "paused"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "default_queue_id"),
					resource.TestCheckResourceAttrSet(singleWorkPoolDatasourceName, "base_job_template"),
				),
			},
			{
				// Check that we can query multiple work pools
				Config: fixtureAccMultipleWorkPools(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(multipleWorkPoolDatasourceName, "work_pools.#", "1"),
					resource.TestCheckResourceAttr(multipleWorkPoolDatasourceName, "work_pools.0.name", "evergreen-pool"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.id"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.created"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.updated"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.type"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.paused"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.default_queue_id"),
					resource.TestCheckResourceAttrSet(multipleWorkPoolDatasourceName, "work_pools.0.base_job_template"),
				),
			},
		},
	})
}
