package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkspaceByHandle(handle string) string {
	return fmt.Sprintf(`
data "prefect_workspace" "evergreen" {
	handle = "%s"
}
`, handle)
}
func fixtureAccWorkspaceByID(id string) string {
	return fmt.Sprintf(`
data "prefect_workspace" "evergreen" {
	id = "%s"
}
`, id)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_workspace(t *testing.T) {
	dataSourceName := "data.prefect_workspace.evergreen"
	workspaceHandle := "evergreen-workspace"
	workspaceID := "45cfa7c6-e136-471c-859b-3be89d0a99ce"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceByHandle(workspaceHandle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "handle", workspaceHandle),
					resource.TestCheckResourceAttr(dataSourceName, "id", workspaceID),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
				),
			},
			{
				Config: fixtureAccWorkspaceByID(workspaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "handle", workspaceHandle),
					resource.TestCheckResourceAttr(dataSourceName, "id", workspaceID),
					resource.TestCheckResourceAttrSet(dataSourceName, "created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "updated"),
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
				),
			},
		},
	})
}
