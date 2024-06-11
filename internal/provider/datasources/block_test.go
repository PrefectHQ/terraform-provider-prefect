package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccBlockByName(name string) string {
	return fmt.Sprintf(`
resource "prefect_block" "%s" {
  name      = "%s"
  type_slug = "secret"

  data = jsonencode({
    "someKey" : "someValue"
  })
}

data "prefect_block" "my_existing_secret_by_id" {
  id = prefect_block.%s.id
}

data "prefect_block" "my_existing_secret_by_name" {
  name            = "%s"
  block_type_name = "secret"
}
`, name, name, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_block(t *testing.T) {
	datasourceNameByID := "data.prefect_block.my_existing_secret_by_id"
	datasourceNameByName := "data.prefect_block.my_existing_secret_by_name"

	blockName := "my_block"
	blockValue := "{\"someKey\":\"someValue\"}"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test block datasource by ID.
				Config: fixtureAccBlockByName(blockName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameByID, "id"),
					resource.TestCheckResourceAttr(datasourceNameByID, "name", blockName),
					resource.TestCheckResourceAttr(datasourceNameByID, "data", blockValue),
				),
			},
			{
				// Test block datasource by name.
				Config: fixtureAccBlockByName(blockName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameByName, "id"),
					resource.TestCheckResourceAttr(datasourceNameByName, "name", blockName),
					resource.TestCheckResourceAttr(datasourceNameByName, "data", blockValue),
				),
			},
		},
	})
}
