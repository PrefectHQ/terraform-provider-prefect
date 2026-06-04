package resources_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

func fixtureAccWorkspaceAccessResourceForBot(workspace, name string) string {
	return fmt.Sprintf(`
%s
data "prefect_workspace_role" "developer" {
	name = "Developer"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = prefect_workspace.test.id
	workspace_role_id = data.prefect_workspace_role.developer.id
}`, workspace, name)
}

func fixtureAccWorkspaceAccessResourceUpdateForBot(workspace, name string) string {
	return fmt.Sprintf(`
%s
data "prefect_workspace_role" "runner" {
	name = "Runner"
}
resource "prefect_service_account" "bot" {
	name = "%s"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = prefect_workspace.test.id
	workspace_role_id = data.prefect_workspace_role.runner.id
}`, workspace, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_bot_workspace_access(t *testing.T) {
	// Workspace access is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	accessResourceName := "prefect_workspace_access.bot_access"
	botResourceName := "prefect_service_account.bot"
	developerRoleDatsourceName := "data.prefect_workspace_role.developer"
	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"
	workspace := testutils.NewEphemeralWorkspace()
	randomName := testutils.NewRandomPrefixedString()

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspaceAccess api.WorkspaceAccess

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResourceForBot(workspace.Resource, randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.ServiceAccount, accessResourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.ServiceAccount, &workspaceAccess, botResourceName, developerRoleDatsourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.CompareValuePairs(accessResourceName, "accessor_id", botResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", developerRoleDatsourceName, "id"),
				},
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForBot(workspace.Resource, randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.ServiceAccount, accessResourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.ServiceAccount, &workspaceAccess, botResourceName, runnerRoleDatsourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.CompareValuePairs(accessResourceName, "accessor_id", botResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				},
			},
			{
				ImportState:       true,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(accessResourceName),
				ResourceName:      accessResourceName,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile(".*Import not supported.*"),
			},
		},
	})
}

func fixtureAccWorkspaceAccessResourceForTeam(workspace string) string {
	return fmt.Sprintf(`
%s
data "prefect_workspace_role" "viewer" {
	name = "Viewer"
}
data "prefect_team" "my_team" {
	name = "my-team"
}
resource "prefect_workspace_access" "team_access" {
	accessor_type = "TEAM"
	accessor_id = data.prefect_team.my_team.id
	workspace_id = prefect_workspace.test.id
	workspace_role_id = data.prefect_workspace_role.viewer.id
}`, workspace)
}

func fixtureAccWorkspaceAccessResourceUpdateForTeam(workspace string) string {
	return fmt.Sprintf(`
%s
data "prefect_workspace_role" "runner" {
	name = "Runner"
}
data "prefect_team" "my_team" {
	name = "my-team"
}
resource "prefect_workspace_access" "team_access" {
	accessor_type = "TEAM"
	accessor_id = data.prefect_team.my_team.id
	workspace_id = prefect_workspace.test.id
	workspace_role_id = data.prefect_workspace_role.runner.id
}`, workspace)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_workspace_access(t *testing.T) {
	// Workspace access is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	accessResourceName := "prefect_workspace_access.team_access"
	teamResourceName := "data.prefect_team.my_team"
	viewerRoleDatsourceName := "data.prefect_workspace_role.viewer"
	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"
	workspace := testutils.NewEphemeralWorkspace()

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspaceAccess api.WorkspaceAccess
	// workspaceID is captured during the test so that the PreConfig hook
	// (which does not receive the Terraform state) can build a scoped client
	// to delete the access grant out-of-band.
	var workspaceID uuid.UUID

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkspaceAccessResourceForTeam(workspace.Resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check creation + existence of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.Team, accessResourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.Team, &workspaceAccess, teamResourceName, viewerRoleDatsourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.CompareValuePairs(accessResourceName, "accessor_id", teamResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", viewerRoleDatsourceName, "id"),
				},
			},
			{
				Config: fixtureAccWorkspaceAccessResourceUpdateForTeam(workspace.Resource),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check updating the role of the workspace access resource, with matching linked attributes
					testAccCheckWorkspaceAccessExists(utils.Team, accessResourceName, &workspaceAccess),
					testAccCheckWorkspaceAccessValuesForAccessor(utils.Team, &workspaceAccess, teamResourceName, runnerRoleDatsourceName),
					captureWorkspaceAccessWorkspaceID(&workspaceID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.CompareValuePairs(accessResourceName, "accessor_id", teamResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
				},
			},
			// Delete the TEAM access grant out-of-band and verify Terraform plans
			// a corrective re-create rather than erroring during read. The TEAM
			// accessor fetches via a 200-OK list filter, so a missing grant
			// surfaces as a synthetic not-found that must still drop the resource
			// from state. Regression guard for PLA-2779.
			{
				PreConfig: func() {
					c, err := testutils.NewTestClient()
					if err != nil {
						t.Fatalf("failed to construct test client: %v", err)
					}
					workspaceAccessClient, err := c.WorkspaceAccess(uuid.Nil, workspaceID)
					if err != nil {
						t.Fatalf("failed to construct workspace access client: %v", err)
					}
					// For TEAM access, Delete is keyed off the accessor (team) ID.
					if err := workspaceAccessClient.Delete(context.Background(), utils.Team, workspaceAccess.ID, *workspaceAccess.TeamID); err != nil {
						t.Fatalf("failed to delete team workspace access out-of-band: %v", err)
					}
				},
				Config:             fixtureAccWorkspaceAccessResourceUpdateForTeam(workspace.Resource),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ImportState:       true,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(accessResourceName),
				ResourceName:      accessResourceName,
				ExpectError:       regexp.MustCompile(".*Import not supported.*"),
			},
		},
	})
}

// captureWorkspaceAccessWorkspaceID stores the ephemeral workspace ID from
// state so a subsequent step's PreConfig hook (which does not receive the
// Terraform state) can build a scoped client to mutate access out-of-band.
func captureWorkspaceAccessWorkspaceID(out *uuid.UUID) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceID, err := testutils.GetResourceWorkspaceIDFromState(state)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}
		*out = workspaceID

		return nil
	}
}

func testAccCheckWorkspaceAccessExists(accessorType string, accessResourceName string, access *api.WorkspaceAccess) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceAccessID, err := testutils.GetResourceIDFromState(state, accessResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace access ID: %w", err)
		}

		workspaceID, err := testutils.GetResourceWorkspaceIDFromState(state)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

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
