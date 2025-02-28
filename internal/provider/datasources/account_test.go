package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(datasourceName, "id", os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")),
					testutils.ExpectKnownValueNotNull(datasourceName, "name"),
					testutils.ExpectKnownValueNotNull(datasourceName, "handle"),

					// These domain names were manually added to the account, because we're using a pre-existing account
					// due to the fact that accounts cannot be created with the API/Terraform.
					testutils.ExpectKnownValueList(datasourceName, "domain_names", []string{"example.com", "foobar.com"}),
				},
			},
		},
	})
}
