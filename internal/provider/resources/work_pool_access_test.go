package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type workPoolAccessConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
}

func fixtureAccWorkPoolAccess(cfg workPoolAccessConfig) string {
	tmpl := `
{{.WorkspaceResource}}

data "prefect_workspace_role" "developer" {
	name = "Developer"
}

resource "prefect_service_account" "test" {
	name = "my-service-account"
}

resource "prefect_workspace_access" "test" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.test.id
	workspace_role_id = data.prefect_workspace_role.developer.id
	workspace_id = {{.WorkspaceResourceName}}.id
}

data "prefect_team" "test" {
	name = "my-team"
}

resource "prefect_workspace_access" "test_team" {
	accessor_type = "TEAM"
	accessor_id = data.prefect_team.test.id
	workspace_role_id = data.prefect_workspace_role.developer.id
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_flow" "test" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_work_pool" "test" {
	name = "my-work-pool"
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_work_pool_access" "test" {
	workspace_id = {{.WorkspaceResourceName}}.id
	work_pool_name = prefect_work_pool.test.name

	manage_actor_ids = [prefect_service_account.test.actor_id]
	run_actor_ids = [prefect_service_account.test.actor_id]
	view_actor_ids = [prefect_service_account.test.actor_id]

	manage_team_ids = [data.prefect_team.test.id]
	run_team_ids = [data.prefect_team.test.id]
	view_team_ids = [data.prefect_team.test.id]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_pool_access(t *testing.T) {
	// Work pool access is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	workspace := testutils.NewEphemeralWorkspace()
	serviceAccountName := "my-service-account"
	teamName := "my-team"

	cfgSet := workPoolAccessConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
	}

	var workPool api.WorkPool
	var workPoolAccess api.WorkPoolAccessControl

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkPoolAccess(cfgSet),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkPoolExists("prefect_work_pool.test", &workPool),
					testAccCheckWorkPoolAccessExists("prefect_work_pool_access.test", &workPoolAccess),
					testAccCheckWorkPoolAccessValues(&workPoolAccess, expectedWorkPoolAccessValues{
						manageActors: []api.ObjectActorAccess{
							{Name: serviceAccountName, Type: api.ServiceAccountAccessor},
							{Name: teamName, Type: api.TeamAccessor},
						},
						runActors: []api.ObjectActorAccess{
							{Name: serviceAccountName, Type: api.ServiceAccountAccessor},
							{Name: teamName, Type: api.TeamAccessor},
						},
						viewActors: []api.ObjectActorAccess{
							{Name: serviceAccountName, Type: api.ServiceAccountAccessor},
							{Name: teamName, Type: api.TeamAccessor},
						},
					}),
				),
			},
		},
	})
}

// testAccCheckWorkPoolAccessExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckWorkPoolAccessExists(workPoolAccessResourceName string, workPoolAccess *api.WorkPoolAccessControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the work pool access resource we just created from the state
		workPoolName, err := testutils.GetResourceAttributeFromStateByAttribute(s, workPoolAccessResourceName, "work_pool_name")
		if err != nil {
			return fmt.Errorf("error fetching work pool name: %w", err)
		}

		// Get the workspace resource we just created from the state
		workspaceID, err := testutils.GetResourceIDFromState(s, testutils.WorkspaceResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		workPoolAccessClient, _ := c.WorkPoolAccess(uuid.Nil, workspaceID)

		fetchedWorkPoolAccess, err := workPoolAccessClient.Read(context.Background(), workPoolName)
		if err != nil {
			return fmt.Errorf("error fetching work pool access: %w", err)
		}

		// Assign the fetched work pool access to the passed pointer
		// so we can use it in the next test assertion
		*workPoolAccess = *fetchedWorkPoolAccess

		return nil
	}
}

type expectedWorkPoolAccessValues struct {
	manageActors []api.ObjectActorAccess
	runActors    []api.ObjectActorAccess
	viewActors   []api.ObjectActorAccess
}

// testAccCheckWorkPoolAccessValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckWorkPoolAccessValues(fetchedWorkPoolAccess *api.WorkPoolAccessControl, expectedValues expectedWorkPoolAccessValues) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		tests := map[string]struct {
			fetched  []api.ObjectActorAccess
			expected []api.ObjectActorAccess
		}{
			"manageActors": {fetchedWorkPoolAccess.ManageActors, expectedValues.manageActors},
			"runActors":    {fetchedWorkPoolAccess.RunActors, expectedValues.runActors},
			"viewActors":   {fetchedWorkPoolAccess.ViewActors, expectedValues.viewActors},
		}

		for name, test := range tests {
			err := actorFound(test.fetched, test.expected)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}

		return nil
	}
}
