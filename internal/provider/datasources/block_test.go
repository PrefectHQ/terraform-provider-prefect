package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccBlockByName(workspace, name string) string {
	aID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")

	return fmt.Sprintf(`
%s

resource "prefect_block" "%s" {
  name      = "%s"
  type_slug = "secret"

  data = jsonencode({
    "someKey" = "someValue"
  })

  account_id = "%s"
  workspace_id = prefect_workspace.test.id
}

data "prefect_block" "my_existing_secret_by_id" {
  id = prefect_block.%s.id

  account_id = "%s"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_block.%s]
}

data "prefect_block" "my_existing_secret_by_name" {
  name      = "%s"
  type_slug = "secret"

  account_id = "%s"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_block.%s]
}
`, workspace, name, name, aID, name, aID, name, name, aID, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_block(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	datasourceNameByID := "data.prefect_block.my_existing_secret_by_id"
	datasourceNameByName := "data.prefect_block.my_existing_secret_by_name"

	blockName := "my-block"
	blockValue := "{\"someKey\":\"someValue\"}"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test block datasource by ID.
				Config: fixtureAccBlockByName(workspace.Resource, blockName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByID, "id"),
					testutils.ExpectKnownValue(datasourceNameByID, "name", blockName),
					testutils.ExpectKnownValue(datasourceNameByID, "data", blockValue),
				},
			},
			{
				// Test block datasource by name.
				Config: fixtureAccBlockByName(workspace.Resource, blockName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByName, "id"),
					testutils.ExpectKnownValue(datasourceNameByName, "name", blockName),
					testutils.ExpectKnownValue(datasourceNameByName, "data", blockValue),
				},
			},
		},
	})
}
