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

const workspaceDatsourceName = "data.prefect_workspace.evergreen"

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
					testAccCheckWorkspaceAccessExists(utils.ServiceAccount, accessResourceName, workspaceDatsourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.ServiceAccount, &workspaceAccess, botResourceName, developerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", developerRoleDatsourceName, "id"),
				),
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForBot(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.ServiceAccount, accessResourceName, workspaceDatsourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.ServiceAccount, &workspaceAccess, botResourceName, runnerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", botResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				),
			},
		},
	})
}

func fixtureAccWorkspaceAccessResourceForTeam() string {
	return `
data "prefect_workspace_role" "viewer" {
	name = "Viewer"
}
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
data "prefect_team" "my_team" {
	name = "my-team"
}
resource "prefect_workspace_access" "team_access" {
	accessor_type = "TEAM"
	accessor_id = data.prefect_team.my_team.id
	workspace_id = data.prefect_workspace.evergreen.id
	workspace_role_id = data.prefect_workspace_role.viewer.id
}`
}

func fixtureAccWorkspaceAccessResourceUpdateForTeam() string {
	return `
data "prefect_workspace_role" "runner" {
	name = "Runner"
}
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
data "prefect_team" "my_team" {
	name = "my-team"
}
resource "prefect_workspace_access" "team_access" {
	accessor_type = "TEAM"
	accessor_id = data.prefect_team.my_team.id
	workspace_id = data.prefect_workspace.evergreen.id
	workspace_role_id = data.prefect_workspace_role.runner.id
}`
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_workspace_access(t *testing.T) {
	accessResourceName := "prefect_workspace_access.team_access"
	teamResourceName := "data.prefect_team.my_team"
	viewerRoleDatsourceName := "data.prefect_workspace_role.viewer"
	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspaceAccess api.WorkspaceAccess

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResourceForTeam(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.Team, accessResourceName, workspaceDatsourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.Team, &workspaceAccess, teamResourceName, viewerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", teamResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", viewerRoleDatsourceName, "id"),
				),
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForTeam(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.Team, accessResourceName, workspaceDatsourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.Team, &workspaceAccess, teamResourceName, runnerRoleDatsourceName),
					resource.TestCheckResourceAttrPair(accessResourceName, "accessor_id", teamResourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_id", workspaceDatsourceName, "id"),
					resource.TestCheckResourceAttrPair(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceAccessExists(accessorType string, accessResourceName string, workspaceDatasourceName string, access *api.WorkspaceAccess) resource.TestCheckFunc {
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

func testAccCheckWorkspaceAccessValuesForAccessor(accessorType string, fetchedAccess *api.WorkspaceAccess, accessorResourceName string, roleDatasourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		accessor, exists := state.RootModule().Resources[accessorResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", accessorResourceName)
		}

		switch accessorType {
		case utils.User:
			if fetchedAccess.UserID.String() != accessor.Primary.ID {
				return fmt.Errorf("Expected Workspace Access UserID to be %s, got %s", accessor.Primary.ID, fetchedAccess.UserID.String())
			}
		case utils.ServiceAccount:
			if fetchedAccess.BotID.String() != accessor.Primary.ID {
				return fmt.Errorf("Expected Workspace Access BotID to be %s, got %s", accessor.Primary.ID, fetchedAccess.BotID.String())
			}
		case utils.Team:
			if fetchedAccess.TeamID.String() != accessor.Primary.ID {
				return fmt.Errorf("Expected Workspace Access TeamID to be %s, got %s", accessor.Primary.ID, fetchedAccess.TeamID.String())
			}
		default:
			return fmt.Errorf("Unsupported accessor type: %s", accessorType)
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
