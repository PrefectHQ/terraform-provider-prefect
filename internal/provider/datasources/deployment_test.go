package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccDeploymentByName(workspace string) string {
	aID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")

	return fmt.Sprintf(`
%s

resource "prefect_flow" "test" {
	name = "test"
	tags = ["test"]

	workspace_id = prefect_workspace.test.id
}

resource "prefect_deployment" "test" {
  name = "test"
  flow_id = prefect_flow.test.id

  description = "test description"
  version = "1.2.3"

  account_id = "%s"
  workspace_id = prefect_workspace.test.id
}

data "prefect_deployment" "test_by_id" {
  id = prefect_deployment.test.id

  account_id = "%s"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_deployment.test]
}

data "prefect_deployment" "test_by_name" {
  name = prefect_deployment.test.name

  account_id = "%s"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_deployment.test]
}
`, workspace, aID, aID, aID)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_deployment(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	datasourceNameByID := "data.prefect_deployment.test_by_id"
	datasourceNameByName := "data.prefect_deployment.test_by_name"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentByName(workspace.Resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameByID, "id"),
					resource.TestCheckResourceAttr(datasourceNameByID, "name", "test"),
					resource.TestCheckResourceAttr(datasourceNameByID, "description", "test description"),
					resource.TestCheckResourceAttr(datasourceNameByID, "version", "1.2.3"),
				),
			},
			{
				Config: fixtureAccDeploymentByName(workspace.Resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameByName, "name"),
					resource.TestCheckResourceAttr(datasourceNameByName, "name", "test"),
					resource.TestCheckResourceAttr(datasourceNameByName, "description", "test description"),
					resource.TestCheckResourceAttr(datasourceNameByName, "version", "1.2.3"),
				),
			},
		},
	})
}
