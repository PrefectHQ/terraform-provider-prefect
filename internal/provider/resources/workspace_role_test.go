package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_workspace_role(t *testing.T) {
	resourceName := "prefect_workspace_role.role"
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the workspace role resource
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleResource(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("%s description", randomName)),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scopes.0", "manage_blocks"),
					resource.TestCheckResourceAttr(resourceName, "scopes.1", "see_artifacts"),
				),
			},
			{
				// Check updates for the workspace role resource
				Config: testutils.ProviderConfig + fixtureAccWorkspaceRoleReesourceUpdated(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("description for %s", randomName)),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "scopes.0", "write_workers"),
					resource.TestCheckResourceAttr(resourceName, "scopes.1", "see_variables"),
					resource.TestCheckResourceAttr(resourceName, "scopes.2", "manage_work_queues"),
				),
			},
		},
	})

}

func fixtureAccWorkspaceRoleResource(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace_role" "role" {
	name = "%s"
	description = "%s description"
	scopes = ["manage_blocks", "see_artifacts"]
}`, name, name)
}

func fixtureAccWorkspaceRoleReesourceUpdated(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace_role" "role" {
	name = "%s"
	description = "description for %s"
	scopes = ["write_workers", "see_variables", "manage_work_queues"]
}`, name, name)
}
