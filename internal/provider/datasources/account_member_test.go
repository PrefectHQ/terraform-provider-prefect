package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccAccountMember(email string) string {
	return fmt.Sprintf(`
data "prefect_account_member" "member" {
	email = "%s"
}
	`, email)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_account_member(t *testing.T) {
	// Account members are not available in OSS.
	testutils.SkipTestsIfOSS(t)

	dataSourceName := "data.prefect_account_member.member"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccAccountMember("marvin@prefect.io"),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(dataSourceName, "email", "marvin@prefect.io"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "id"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "account_role_id"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "account_role_name"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "actor_id"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "first_name"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "last_name"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "handle"),
					testutils.ExpectKnownValueNotNull(dataSourceName, "user_id"),
				},
			},
		},
	})
}
