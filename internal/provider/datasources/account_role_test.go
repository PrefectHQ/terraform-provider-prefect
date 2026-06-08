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
	// Account roles are not available in OSS.
	testutils.SkipTestsIfOSS(t)

	dataSourceName := "data.prefect_account_role.test"

	type defaultAccountRole struct {
		name string
		// minPermissionCount is the minimum number of permissions a default
		// role is expected to have. We assert a minimum rather than an exact
		// count because the default permission set varies across environments
		// (for example, Prefect Cloud vs. customer-managed instances).
		minPermissionCount int
	}

	// Default account role names - these exist in every account.
	// The minimums are set conservatively below the lowest observed count per
	// role across environments (Cloud: 44/13/46, customer-managed: 40/11/...),
	// so the test asserts the roles carry a substantial permission set without
	// being brittle to per-environment drift.
	defaultAccountRoles := []defaultAccountRole{{"Admin", 30}, {"Member", 8}, {"Owner", 30}}

	testSteps := make([]resource.TestStep, 0, len(defaultAccountRoles))

	for _, role := range defaultAccountRoles {
		testSteps = append(testSteps, resource.TestStep{
			Config: fixtureAccAccountRoleDataSource(role.name),
			ConfigStateChecks: []statecheck.StateCheck{
				testutils.ExpectKnownValue(dataSourceName, "name", role.name),
				testutils.ExpectKnownValueListSizeMin(dataSourceName, "permissions", role.minPermissionCount),
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
