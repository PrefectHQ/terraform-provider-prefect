package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_service_account(t *testing.T) {
	dataSourceNameByID := "data.prefect_service_account.bot_by_id"
	dataSourceNameByName := "data.prefect_service_account.bot_by_name"
	// generate random resource name
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountDataSource(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the prefect_service_account datasource that queries by ID
					resource.TestCheckResourceAttr(dataSourceNameByID, "name", randomName),
					resource.TestMatchResourceAttr(dataSourceNameByID, "api_key_name", regexp.MustCompile((fmt.Sprintf(`^%s`, randomName)))),
					resource.TestCheckResourceAttrSet(dataSourceNameByID, "created"),
					resource.TestCheckResourceAttrSet(dataSourceNameByID, "updated"),
					// Check the prefect_service_account datasource that queries by name
					resource.TestCheckResourceAttr(dataSourceNameByName, "name", randomName),
					resource.TestMatchResourceAttr(dataSourceNameByName, "api_key_name", regexp.MustCompile((fmt.Sprintf(`^%s`, randomName)))),
					resource.TestCheckResourceAttrSet(dataSourceNameByName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceNameByName, "updated"),
				),
			},
		},
	})
}

func fixtureAccServiceAccountDataSource(name string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
}
data "prefect_service_account" "bot_by_id" {
	id = prefect_service_account.bot.id
}
data "prefect_service_account" "bot_by_name" {
	name = prefect_service_account.bot.name
}
	`, name)
}
