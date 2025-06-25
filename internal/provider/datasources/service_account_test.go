package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_service_account(t *testing.T) {
	// Service accounts are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	dataSourceNameByID := "data.prefect_service_account.bot_by_id"
	dataSourceNameByName := "data.prefect_service_account.bot_by_name"
	randomName := testutils.NewRandomPrefixedString()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountDataSource(randomName),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check the prefect_service_account datasource that queries by ID
					testutils.ExpectKnownValue(dataSourceNameByID, "name", randomName),
					statecheck.ExpectKnownValue(dataSourceNameByID, tfjsonpath.New("api_key_name"), knownvalue.StringRegexp(regexp.MustCompile(fmt.Sprintf(`^%s`, randomName)))),
					testutils.ExpectKnownValueNotNull(dataSourceNameByID, "created"),
					testutils.ExpectKnownValueNotNull(dataSourceNameByID, "updated"),
					testutils.ExpectKnownValueNotNull(dataSourceNameByID, "actor_id"),
					// Check the prefect_service_account datasource that queries by name
					testutils.ExpectKnownValue(dataSourceNameByName, "name", randomName),
					statecheck.ExpectKnownValue(dataSourceNameByName, tfjsonpath.New("api_key_name"), knownvalue.StringRegexp(regexp.MustCompile(fmt.Sprintf(`^%s`, randomName)))),
					testutils.ExpectKnownValueNotNull(dataSourceNameByName, "created"),
					testutils.ExpectKnownValueNotNull(dataSourceNameByName, "updated"),
					testutils.ExpectKnownValueNotNull(dataSourceNameByName, "actor_id"),
				},
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
