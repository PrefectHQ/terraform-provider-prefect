package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccDeploymentAccessCreate(name string) string {
	return fmt.Sprintf(`
data "prefect_account_member" "member" {
	email = "richard@skyscrapr.io"
}

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

resource "prefect_deployment_access" "deployment_access" {
	deployment_id = prefect_deployment.deployment.id
	workspace_id = prefect_workspace.workspace.id
	access_control = {
		manage_actor_ids = [
			data.prefect_account_member.member.actor_id
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
func TestAccResource_deployment_access(t *testing.T) {

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
				Config: fixtureAccDeploymentAccessCreate(randomName),
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

// func testAccCheckDeploymentAccessExists(accessResourceName string, workspaceDatasourceName string, accessorType string, access *api.WorkspaceAccess) resource.TestCheckFunc {
// 	return func(state *terraform.State) error {
// 		workspaceAccessResource, exists := state.RootModule().Resources[accessResourceName]
// 		if !exists {
// 			return fmt.Errorf("Resource not found in state: %s", accessResourceName)
// 		}

// 		workspaceDatsource, exists := state.RootModule().Resources[workspaceDatasourceName]
// 		if !exists {
// 			return fmt.Errorf("Resource not found in state: %s", workspaceDatasourceName)
// 		}

// 		workspaceID, _ := uuid.Parse(workspaceDatsource.Primary.ID)
// 		workspaceAccessID, _ := uuid.Parse(workspaceAccessResource.Primary.ID)

// 		// Create a new client, and use the default accountID from environment
// 		c, _ := testutils.NewTestClient()
// 		workspaceAccessClient, _ := c.WorkspaceAccess(uuid.Nil, workspaceID)

// 		fetchedWorkspaceAccess, err := workspaceAccessClient.Get(context.Background(), accessorType, workspaceAccessID)
// 		if err != nil {
// 			return fmt.Errorf("Error fetching Workspace Access: %w", err)
// 		}
// 		if fetchedWorkspaceAccess == nil {
// 			return fmt.Errorf("Workspace Access not found for ID: %s", workspaceAccessID)
// 		}

// 		*access = *fetchedWorkspaceAccess

// 		return nil
// 	}
// }

// func testAccCheckDeploymentAccessValuesForBot(fetchedAccess *api.WorkspaceAccess, botResourceName string, roleDatasourceName string) resource.TestCheckFunc {
// 	return func(state *terraform.State) error {
// 		bot, exists := state.RootModule().Resources[botResourceName]
// 		if !exists {
// 			return fmt.Errorf("Resource not found in state: %s", botResourceName)
// 		}

// 		if fetchedAccess.BotID.String() != bot.Primary.ID {
// 			return fmt.Errorf("Expected Workspace Access BotID to be %s, got %s", bot.Primary.ID, fetchedAccess.BotID.String())
// 		}

// 		role, exists := state.RootModule().Resources[roleDatasourceName]
// 		if !exists {
// 			return fmt.Errorf("Resource not found in state: %s", roleDatasourceName)
// 		}

// 		if fetchedAccess.WorkspaceRoleID.String() != role.Primary.ID {
// 			return fmt.Errorf("Expected Workspace Access WorkspaceRoleID to be %s, got %s", role.Primary.ID, fetchedAccess.WorkspaceRoleID.String())
// 		}

// 		return nil
// 	}
// }
