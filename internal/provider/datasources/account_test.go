package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccAccount() string {
	return fmt.Sprintf(`
data "prefect_account" "test" {
	id = "%s"
}
	`, os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"))
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_account(t *testing.T) {
	datasourceName := "data.prefect_account.test"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccAccount(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "id", os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")),
					resource.TestCheckResourceAttrSet(datasourceName, "name"),
					resource.TestCheckResourceAttrSet(datasourceName, "handle"),
				),
			},
		},
	})
}
