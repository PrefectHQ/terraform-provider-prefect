package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccAccountRoleDataSource(name string) string {
	return fmt.Sprintf(`
data "prefect_account_role" "test" {
	name = "%s"
}
	`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_account_role_defaults(t *testing.T) {
	dataSourceName := "data.prefect_account_role.test"

	type defaultAccountRole struct {
		name            string
		permissionCount string
	}

	// Default account role names - these exist in every account
	defaultAccountRoles := []defaultAccountRole{{"Admin", "37"}, {"Member", "10"}, {"Owner", "10"}}

	testSteps := []resource.TestStep{}

	for _, role := range defaultAccountRoles {
		testSteps = append(testSteps, resource.TestStep{
			Config: fixtureAccAccountRoleDataSource(role.name),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(dataSourceName, "name", role.name),
				resource.TestCheckResourceAttr(dataSourceName, "permissions.#", role.permissionCount),
				// Default roles should not be associated with an account
				resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
			),
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps:                    testSteps,
	})
}
