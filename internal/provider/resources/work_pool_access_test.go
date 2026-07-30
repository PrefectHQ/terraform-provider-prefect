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
	ServiceAccountName    string
	IncludeAccess         bool
}

func fixtureAccWorkPoolAccess(cfg workPoolAccessConfig) string {
	tmpl := `
{{.WorkspaceResource}}

data "prefect_workspace_role" "developer" {
	name = "Developer"
}

resource "prefect_service_account" "test" {
	name = "{{.ServiceAccountName}}"
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

{{if .IncludeAccess}}
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
{{end}}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_pool_access(t *testing.T) {
	// Work pool access is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	workspace := testutils.NewEphemeralWorkspace()
	serviceAccountName := testutils.NewRandomPrefixedString()
	teamName := "my-team"

	baseCfg := workPoolAccessConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
		ServiceAccountName:    serviceAccountName,
	}

	var workPool api.WorkPool
	var workPoolAccess api.WorkPoolAccessControl

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccWorkPoolAccess(workPoolAccessConfig{
					WorkspaceResource:     baseCfg.WorkspaceResource,
					WorkspaceResourceName: baseCfg.WorkspaceResourceName,
					ServiceAccountName:    baseCfg.ServiceAccountName,
					IncludeAccess:         true,
				}),
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
			{
				Config: fixtureAccWorkPoolAccess(baseCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkPoolAccessDestroy("prefect_work_pool.test"),
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

// testAccCheckWorkPoolAccessDestroy is a Custom Check Function that
// verifies that the API object was destroyed correctly.
func testAccCheckWorkPoolAccessDestroy(workPoolResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the workspace resource we just created from the state
		workspaceID, err := testutils.GetResourceIDFromState(s, testutils.WorkspaceResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		workPoolName, err := testutils.GetResourceAttributeFromStateByAttribute(s, workPoolResourceName, "name")
		if err != nil {
			return fmt.Errorf("error fetching work pool name: %w", err)
		}

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		workPoolAccessClient, _ := c.WorkPoolAccess(uuid.Nil, workspaceID)

		fetchedWorkPoolAccess, err := workPoolAccessClient.Read(context.Background(), workPoolName)
		if err != nil {
			return fmt.Errorf("error fetching work pool access: %w", err)
		}

		expectedActors := []api.ObjectActorAccess{
			{ID: "*", Name: "*", Type: api.AllAccessors},
		}

		checks := map[string][]api.ObjectActorAccess{
			"manage_actors": fetchedWorkPoolAccess.ManageActors,
			"run_actors":    fetchedWorkPoolAccess.RunActors,
			"view_actors":   fetchedWorkPoolAccess.ViewActors,
		}
		for name, actors := range checks {
			if len(actors) != len(expectedActors) {
				return fmt.Errorf("expected %s to have %d entries, got %d", name, len(expectedActors), len(actors))
			}
			if actors[0].ID != "*" || actors[0].Type != api.AllAccessors {
				return fmt.Errorf("expected %s to be wildcard access, got %+v", name, actors[0])
			}
		}

		return nil
	}
}
