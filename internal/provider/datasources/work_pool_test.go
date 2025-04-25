package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccSingleWorkPool(workspace, workspaceIDArg, name string) string {
	return fmt.Sprintf(`
%s

resource "prefect_work_pool" "test" {
	name = "%s"
	type = "kubernetes"
	%s
}

data "prefect_work_pool" "test" {
	name = "%s"
	%s
	depends_on = [prefect_work_pool.test]
}
`, workspace, name, workspaceIDArg, name, workspaceIDArg)
}

func fixtureAccMultipleWorkPools(workspace, workspaceIDArg, name string) string {
	return fmt.Sprintf(`
%s

resource "prefect_work_pool" "test" {
	name = "%s"
	type = "kubernetes"
	%s
}

data "prefect_work_pools" "test" {
	%s
	depends_on = [prefect_work_pool.test]
}
`, workspace, name, workspaceIDArg, workspaceIDArg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_work_pool(t *testing.T) {
	singleWorkPoolDatasourceName := "data.prefect_work_pool.test"
	multipleWorkPoolDatasourceName := "data.prefect_work_pools.test"
	workspace := testutils.NewEphemeralWorkspace()
	workPoolName := testutils.NewRandomPrefixedString()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check that we can query a single work pool
				Config: fixtureAccSingleWorkPool(workspace.Resource, workspace.IDArg, workPoolName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(singleWorkPoolDatasourceName, "name", workPoolName),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "id"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "created"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "updated"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "type"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "paused"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "default_queue_id"),
					testutils.ExpectKnownValueNotNull(singleWorkPoolDatasourceName, "base_job_template"),
				},
			},
			{
				// Check that we can query multiple work pools
				Config: fixtureAccMultipleWorkPools(workspace.Resource, workspace.IDArg, workPoolName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(multipleWorkPoolDatasourceName, "work_pools.0.name", workPoolName),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.id"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.created"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.updated"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.type"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.paused"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.default_queue_id"),
					testutils.ExpectKnownValueNotNull(multipleWorkPoolDatasourceName, "work_pools.0.base_job_template"),
				},
			},
		},
	})
}
