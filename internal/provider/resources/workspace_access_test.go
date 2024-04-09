package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

func fixtureAccWorkspaceAccessResourceForBot(botName string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "developer" {
	name = "Developer"
}
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
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

func fixtureAccWorkspaceAccessResourceUpdateForBot(botName string) string {
	return fmt.Sprintf(`
data "prefect_workspace_role" "runner" {
	name = "Runner"
}
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = data.prefect_workspace.evergreen.id
	workspace_role_id = data.prefect_workspace_role.runner.id
}`, botName)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_bot_workspace_access(t *testing.T) {
	accessResourceName := "prefect_workspace_access.bot_access"
	botResourceName := "prefect_service_account.bot"
	workspaceDatsourceName := "data.prefect_workspace.evergreen"
	developerRoleDatsourceName := "data.prefect_workspace_role.developer"
	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"

	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspaceAccess api.WorkspaceAccess

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResourceForBot(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(accessResourceName, workspaceDatsourceName, utils.ServiceAccount, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForBot(&workspaceAccess, botResourceName, developerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", developerRoleDatsourceName, "id"),
				),
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForBot(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(accessResourceName, workspaceDatsourceName, utils.ServiceAccount, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForBot(&workspaceAccess, botResourceName, runnerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceAccessExists(accessResourceName string, workspaceDatasourceName string, accessorType string, access *api.WorkspaceAccess) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceAccessResource, exists := state.RootModule().Resources[accessResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", accessResourceName)
		}

		workspaceDatsource, exists := state.RootModule().Resources[workspaceDatasourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workspaceDatasourceName)
		}

		workspaceID, _ := uuid.Parse(workspaceDatsource.Primary.ID)
		workspaceAccessID, _ := uuid.Parse(workspaceAccessResource.Primary.ID)

		// Create a new client, and use the default accountID from environment
		c, _ := testutils.NewTestClient()
		workspaceAccessClient, _ := c.WorkspaceAccess(uuid.Nil, workspaceID)

		fetchedWorkspaceAccess, err := workspaceAccessClient.Get(context.Background(), accessorType, workspaceAccessID)
		if err != nil {
			return fmt.Errorf("Error fetching Workspace Access: %w", err)
		}
		if fetchedWorkspaceAccess == nil {
			return fmt.Errorf("Workspace Access not found for ID: %s", workspaceAccessID)
		}

		*access = *fetchedWorkspaceAccess

		return nil
	}
}

func testAccCheckWorkspaceAccessValuesForBot(fetchedAccess *api.WorkspaceAccess, botResourceName string, roleDatasourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		bot, exists := state.RootModule().Resources[botResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", botResourceName)
		}

		if fetchedAccess.BotID.String() != bot.Primary.ID {
			return fmt.Errorf("Expected Workspace Access BotID to be %s, got %s", bot.Primary.ID, fetchedAccess.BotID.String())
		}

		role, exists := state.RootModule().Resources[roleDatasourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", roleDatasourceName)
		}

		if fetchedAccess.WorkspaceRoleID.String() != role.Primary.ID {
			return fmt.Errorf("Expected Workspace Access WorkspaceRoleID to be %s, got %s", role.Primary.ID, fetchedAccess.WorkspaceRoleID.String())
		}

		return nil
	}
}
