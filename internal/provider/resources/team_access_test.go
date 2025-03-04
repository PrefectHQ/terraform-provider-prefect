package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccTeamAccessResourceForServiceAccount(name string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "test" {
	name = "%s"
}

resource "prefect_team" "test" {
	name = "%s"
	description = "test-team-description"
}

resource "prefect_team_access" "test" {
	member_type = "service_account"
	member_id = prefect_service_account.test.id
	member_actor_id = prefect_service_account.test.actor_id
	team_id = prefect_team.test.id
}
`, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_access(t *testing.T) {
	accessResourceName := "prefect_team_access.test"
	randomName := testutils.NewRandomPrefixedString()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccTeamAccessResourceForServiceAccount(randomName),
				Check:  resource.ComposeAggregateTestCheckFunc(),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(accessResourceName, "member_type", "service_account"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "member_id"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "team_id"),
				},
			},
		},
	})
}

// func fixtureAccWorkspaceAccessResourceForTeam(workspace string) string {
// 	return fmt.Sprintf(`
// %s
// data "prefect_workspace_role" "viewer" {
// 	name = "Viewer"
// }
// data "prefect_team" "my_team" {
// 	name = "my-team"
// }
// resource "prefect_workspace_access" "team_access" {
// 	accessor_type = "TEAM"
// 	accessor_id = data.prefect_team.my_team.id
// 	workspace_id = prefect_workspace.test.id
// 	workspace_role_id = data.prefect_workspace_role.viewer.id
// }`, workspace)
// }

// func fixtureAccWorkspaceAccessResourceUpdateForTeam(workspace string) string {
// 	return fmt.Sprintf(`
// %s
// data "prefect_workspace_role" "runner" {
// 	name = "Runner"
// }
// data "prefect_team" "my_team" {
// 	name = "my-team"
// }
// resource "prefect_workspace_access" "team_access" {
// 	accessor_type = "TEAM"
// 	accessor_id = data.prefect_team.my_team.id
// 	workspace_id = prefect_workspace.test.id
// 	workspace_role_id = data.prefect_workspace_role.runner.id
// }`, workspace)
// }

// //nolint:paralleltest // we use the resource.ParallelTest helper instead
// func TestAccResource_team_workspace_access(t *testing.T) {
// 	accessResourceName := "prefect_workspace_access.team_access"
// 	teamResourceName := "data.prefect_team.my_team"
// 	viewerRoleDatsourceName := "data.prefect_workspace_role.viewer"
// 	runnerRoleDatsourceName := "data.prefect_workspace_role.runner"
// 	workspace := testutils.NewEphemeralWorkspace()

// 	// We use this variable to store the fetched resource from the API
// 	// and it will be shared between TestSteps via a pointer.
// 	var workspaceAccess api.WorkspaceAccess

// 	resource.ParallelTest(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
// 		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
// 		Steps: []resource.TestStep{
// 			{
// 				Config: fixtureAccWorkspaceAccessResourceForTeam(workspace.Resource),
// 				Check:  resource.ComposeAggregateTestCheckFunc(),
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					testutils.CompareValuePairs(accessResourceName, "accessor_id", teamResourceName, "id"),
// 					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
// 					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", viewerRoleDatsourceName, "id"),
// 				},
// 			},
// 			{
// 				Config: fixtureAccWorkspaceAccessResourceUpdateForTeam(workspace.Resource),
// 				Check:  resource.ComposeAggregateTestCheckFunc(),
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					testutils.CompareValuePairs(accessResourceName, "accessor_id", teamResourceName, "id"),
// 					testutils.CompareValuePairs(accessResourceName, "workspace_id", testutils.WorkspaceResourceName, "id"),
// 					testutils.CompareValuePairs(accessResourceName, "workspace_role_id", runnerRoleDatsourceName, "id"),
// 				},
// 			},
// 		},
// 	})
// }
