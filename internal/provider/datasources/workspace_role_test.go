package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func TestWorkspaceRolesDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testutils.ProviderConfig + `
data "prefect_workspace_role" "test" {
	name = "Owner"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.prefect_workspace_role.test", "name", "Owner"),
				),
			},
			{
				Config: testutils.ProviderConfig + `
data "prefect_workspace_role" "test" {
	name = "Worker"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.prefect_workspace_role.test", "name", "Worker"),
				),
			},
			{
				Config: testutils.ProviderConfig + `
data "prefect_workspace_role" "test" {
	name = "Developer"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.prefect_workspace_role.test", "name", "Developer"),
				),
			},
			{
				Config: testutils.ProviderConfig + `
data "prefect_workspace_role" "test" {
	name = "Viewer"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.prefect_workspace_role.test", "name", "Viewer"),
				),
			},
			{
				Config: testutils.ProviderConfig + `
data "prefect_workspace_role" "test" {
	name = "Runner"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.prefect_workspace_role.test", "name", "Runner"),
				),
			},
		},
	})
}
