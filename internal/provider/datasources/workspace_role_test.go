package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkspaceRoleDataSource(name string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "test" {
	name = "%s"
}
	`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_workspace_role_defaults(t *testing.T) {
	dataSourceName := "data.prefect_workspace_role.test"

	// Default workspace role names - these exist in every account
	defaultWorkspaceRoles := []string{"Owner", "Worker", "Developer", "Viewer", "Runner"}

	testSteps := []resource.TestStep{}

	for _, role := range defaultWorkspaceRoles {
		testSteps = append(testSteps, resource.TestStep{
			Config: fixtureAccWorkspaceRoleDataSource(role),
			ConfigStateChecks: []statecheck.StateCheck{
				testutils.ExpectKnownValue(dataSourceName, "name", role),
				testutils.ExpectKnownValueNotNull(dataSourceName, "created"),
				testutils.ExpectKnownValueNotNull(dataSourceName, "updated"),
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
