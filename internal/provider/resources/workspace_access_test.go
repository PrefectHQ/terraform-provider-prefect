package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkspaceAccessResource(botName string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "developer" {
	name = "Developer"
}
data "prefect_workspace" "evergreen" {
	id = "45cfa7c6-e136-471c-859b-3be89d0a99ce"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = data.prefect_workspace.evergreen.id
	workspace_role_id = data.prefect_workspace_role.developer.id
}`, botName)
}

func TestAccResource_bot_workspace_access(t *testing.T) {
	resourceName := "prefect_workspace_access.bot_access"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResource(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					resource.TestCheckResourceAttrPair(resourceName, "accessor_id", "prefect_service_account.bot", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", "data.prefect_workspace.evergreen", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_role_id", "data.prefect_workspace_role.developer", "id"),
				),
			},
		},
	})
}
