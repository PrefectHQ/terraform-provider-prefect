package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_workspace_role_defaults(t *testing.T) {
	dataSourceName := "data.prefect_workspace_role.test"
	// Default workspace role names - these exist in every account
	owner := "Owner"
	worker := "Worker"
	developer := "Developer"
	viewer := "Viewer"
	runner := "Runner"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleDataSource(owner),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", owner),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					// Default roles should not be associated with an account
					resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
				),
			},
			{
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleDataSource(worker),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", worker),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					// Default roles should not be associated with an account
					resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
				),
			},
			{
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleDataSource(developer),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", developer),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					// Default roles should not be associated with an account
					resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
				),
			},
			{
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleDataSource(viewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", viewer),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					// Default roles should not be associated with an account
					resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
				),
			},
			{
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleDataSource(runner),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", runner),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					// Default roles should not be associated with an account
					resource.TestCheckNoResourceAttr(dataSourceName, "account_id"),
				),
			},
		},
	})
}

func fixtureAccWorkspaceRoleDataSource(name string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "test" {
	name = "%s"
}
	`, name)
}
