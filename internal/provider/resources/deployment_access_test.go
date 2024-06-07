package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccDeploymentAccess_members(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}

resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = prefect_workspace.workspace.id
	tags = ["test"]
}

resource "prefect_deployment" "deployment" {
	name = "%s"
	description = "string"
	workspace_id = prefect_workspace.workspace.id
	flow_id = prefect_flow.flow.id
	entrypoint = "hello_world.py:hello_world"
	tags = ["test"]
}

data prefect_account_members account_members {}

resource "prefect_deployment_access" "deployment_access" {
	deployment_id = prefect_deployment.deployment.id
	workspace_id = prefect_workspace.workspace.id
	access_control = {
		manage_actor_ids = [
			data.prefect_account_members.account_members.members[0].actor_id
		]
	}
}
`, name, name, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_access_members(t *testing.T) {
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resourceName := "prefect_deployment_access.deployment_access"
	deploymentResourceName := "prefect_deployment.deployment"
	workspaceResourceName := "prefect_workspace.workspace"
	accountMembersDatasourcename := "data.prefect_account_members.account_members"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentAccess_members(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", deploymentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_id", workspaceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_id", deploymentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "manage_actor_ids.0", accountMembersDatasourcename, "members.0.actor_id"),
					resource.TestCheckResourceAttrPair(resourceName, "result.manage_actors.0.id", accountMembersDatasourcename, "members.0.actor_id"),
				),
			},
			// {
			// 	Config: fixtureAccWorkspaceAccessResourceUpdateForBot(randomName),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttrSet(accessResourceName, "id"),
			// 		// Check updating the role of the workspace access resource, with matching linked attributes
			// 		// testAccCheckWorkspaceAccessExists(accessResourceName, workspaceDatsourceName, utils.ServiceAccount, &workspaceAccess),
			// 		// testAccCheckWorkspaceAccessValuesForBot(&workspaceAccess, botResourceName, runnerRoleDatsourceName),
			// 		// resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
			// 		// resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
			// 		// resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
			// 	),
			// },
		},
	})
}
