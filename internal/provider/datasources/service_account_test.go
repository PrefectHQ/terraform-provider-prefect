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
	dataSourceName := "data.prefect_service_account.bot"
	// generate random resource name
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountDataSource(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the prefect_service_account datasource
					resource.TestCheckResourceAttr(dataSourceName, "name", randomName),
					resource.TestMatchResourceAttr(dataSourceName, "api_key_name", regexp.MustCompile((fmt.Sprintf(`^%s`, randomName)))),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
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
data "prefect_service_account" "bot" {
	id = prefect_service_account.bot.id
}
	`, name)
}
