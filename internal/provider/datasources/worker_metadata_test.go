package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkerMetadtata() string {
	aID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")

	return fmt.Sprintf(`
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}

data "prefect_worker_metadata" "default" {
  account_id = "%s"
  workspace_id = data.prefect_workspace.evergreen.id
}
`, aID)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_worker_metadata(t *testing.T) {
	datasourceName := "data.prefect_worker_metadata.default"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkerMetadtata(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "base_job_configs.%", "13"),
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
					resource.TestCheckResourceAttrSet(datasourceName, "base_job_configs.ecs_push"),
				),
			},
		}})
}
