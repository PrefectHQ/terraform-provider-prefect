package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkerMetadtata(workspace string) string {
	aID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")

	return fmt.Sprintf(`
%s

data "prefect_worker_metadata" "default" {
  account_id = "%s"
  workspace_id = prefect_workspace.test.id
}
`, workspace, aID)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_worker_metadata(t *testing.T) {
	datasourceName := "data.prefect_worker_metadata.default"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkerMetadtata(workspace.Resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "base_job_configs.%", "14"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.kubernetes"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.ecs"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.azure_container_instances"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.docker"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.cloud_run"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.cloud_run_v2"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.vertex_ai"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.prefect_agent"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.process"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.azure_container_instances_push"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.cloud_run_push"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.cloud_run_v2_push"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.modal_push"),
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.ecs_push"),
				),
			},
		}})
}
