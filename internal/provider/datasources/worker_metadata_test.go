package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkerMetadtata(workspace, workspaceIDArg string) string {
	return fmt.Sprintf(`
%s

data "prefect_worker_metadata" "default" {
	%s
}
`, workspace, workspaceIDArg)
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
				Config: fixtureAccWorkerMetadtata(workspace.Resource, workspace.IDArg),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						datasourceName,
						tfjsonpath.New("base_job_configs"),
						knownvalue.MapSizeExact(15),
					),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.kubernetes"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.ecs"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.azure_container_instances"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.docker"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.cloud_run"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.cloud_run_v2"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.vertex_ai"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.prefect_agent"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.process"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.azure_container_instances_push"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.cloud_run_push"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.cloud_run_v2_push"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.modal_push"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.ecs_push"),
					testutils.ExpectKnownValueNotNull(datasourceName, "base_job_configs.prefect_managed"),
				},
			},
		}})
}
