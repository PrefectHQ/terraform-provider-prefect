package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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
		permissionCount int
	}

	// Default account role names - these exist in every account
	defaultAccountRoles := []defaultAccountRole{{"Admin", 44}, {"Member", 13}, {"Owner", 46}}

	testSteps := []resource.TestStep{}

	for _, role := range defaultAccountRoles {
		testSteps = append(testSteps, resource.TestStep{
			Config: fixtureAccAccountRoleDataSource(role.name),
			ConfigStateChecks: []statecheck.StateCheck{
				testutils.ExpectKnownValue(dataSourceName, "name", role.name),
				testutils.ExpectKnownValueListSize(dataSourceName, "permissions", role.permissionCount),
				// Default roles should not be associated with an account
				testutils.ExpectKnownValueNull(dataSourceName, "account_id"),
			},
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps:                    testSteps,
	})
}
