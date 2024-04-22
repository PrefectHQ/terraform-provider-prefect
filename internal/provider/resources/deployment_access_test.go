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

	// "schedules": [
	// 	{
	// 		"active": true,
	// 		"schedule": {
	// 			"interval": 0,
	// 			"anchor_date": "2019-08-24T14:15:22Z",
	// 			"timezone": "America/New_York"
	// 		}
	// 	}
	// ],
	// parameters: {
	// 	'goodbye': True
	// },
	// parameter_openapi_schema = {
	// 	'title': 'Parameters',
	// 	'type': 'object',
	// 	'properties': {
	// 		'name': {...},
	// 		'goodbye': {...}
	// 	}
	// }

	// is_schedule_active = true
	// paused = false
	// 
	// enforce_parameter_schema = false
	// "parameter_openapi_schema": { },
	// "pull_steps": [
	// { }
	// ],
	
	// manifest_path = "string"
	// work_queue_name = "string"
	// work_pool_name = "my-work-pool"
	// storage_document_id = "e0212ac4-7ec3-401b-b1e6-2a4627d8e7ec"
	// infrastructure_document_id = "ce9a08a7-d77b-4b3f-826a-53820cfe01fa"
	// schedule = {
	// 	interval = 0
	// 	anchor_date = "2019-08-24T14:15:22Z"
	// 	timezone = "America/New_York"
	// },
	// path = "string"
	// version = "string"
	// infra_overrides = { }
}

data prefect_account_members account_members {}

resource "prefect_deployment_access" "deployment_access" {
	deployment_id = prefect_deployment.deployment.id
	workspace_id = prefect_workspace.workspace.id
	access_control = {
		manage_actor_ids = [
			data.prefect_account_members.account_members.members[0].actor_id
		]
		// run_actor_ids = [
		// 	data.prefect_account_member.member.id
		// ]
		// view_actor_ids = [
		// 	data.prefect_account_member.member.id
		// ]
		// manage_team_ids = [
		// 	"497f6eca-6276-4993-bfeb-53cbbbba6f08"
		// ]
		// run_team_ids = [
		// 	"497f6eca-6276-4993-bfeb-53cbbbba6f08"
		// ]
		// view_team_ids = [
		// 	"497f6eca-6276-4993-bfeb-53cbbbba6f08"
		// ]
	}
}
`, name, name, name, name)
}

// func fixtureAccDeploymentAccessResourceUpdateForBot(botName string) string {
// 	return fmt.Sprintf(`
// data "prefect_workspace_role" "runner" {
// 	name = "Runner"
// }
// data "prefect_workspace" "evergreen" {
// 	handle = "evergreen-workspace"
// }
// resource "prefect_service_account" "bot" {
// 	name = "%s"
// }
// resource "prefect_workspace_access" "bot_access" {
// 	accessor_type = "SERVICE_ACCOUNT"
// 	accessor_id = prefect_service_account.bot.id
// 	workspace_id = data.prefect_workspace.evergreen.id
// 	workspace_role_id = data.prefect_workspace_role.runner.id
// }`, botName)
// }

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_access_members(t *testing.T) {
	// resourceName := "prefect_deployment.deployment"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	accessResourceName := "prefect_deployment_access.deployment_access"
	// botResourceName := "prefect_service_account.bot"
	// workspaceDatsourceName := "data.prefect_workspace.evergreen"
	// developerRoleDatsourceName := "data.prefect_workspace_role.developer"
	// runnerRoleDatsourceName := "data.prefect_workspace_role.runner"

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	// var workspaceAccess api.WorkspaceAccess

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentAccess_members(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(accessResourceName, "id"),
					// resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					// resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", developerRoleDatsourceName, "id"),
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
