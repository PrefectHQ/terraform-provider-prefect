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

type blockAccessConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
	ServiceAccountName    string
	BlockName             string
	IncludeAccess         bool
}

func fixtureAccBlockAccess(cfg blockAccessConfig) string {
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

resource "prefect_block" "test" {
	name = "{{.BlockName}}"
	type_slug = "secret"
	data = jsonencode({
		"value" = "test-value"
	})
	workspace_id = {{.WorkspaceResourceName}}.id
}

{{if .IncludeAccess}}
resource "prefect_block_access" "test" {
	workspace_id = {{.WorkspaceResourceName}}.id
	block_id = prefect_block.test.id

	manage_actor_ids = [prefect_service_account.test.actor_id]
	view_actor_ids = [prefect_service_account.test.actor_id]

	manage_team_ids = [data.prefect_team.test.id]
	view_team_ids = [data.prefect_team.test.id]
}
{{end}}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block_access(t *testing.T) {
	// Block access is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	workspace := testutils.NewEphemeralWorkspace()
	serviceAccountName := testutils.NewRandomPrefixedString()
	blockName := testutils.NewRandomPrefixedString()
	teamName := "my-team"

	baseCfg := blockAccessConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
		ServiceAccountName:    serviceAccountName,
		BlockName:             blockName,
	}

	var blockAccess api.BlockDocumentAccess

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccBlockAccess(blockAccessConfig{
					WorkspaceResource:     baseCfg.WorkspaceResource,
					WorkspaceResourceName: baseCfg.WorkspaceResourceName,
					ServiceAccountName:    baseCfg.ServiceAccountName,
					BlockName:             baseCfg.BlockName,
					IncludeAccess:         true,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockAccessExists("prefect_block_access.test", &blockAccess),
					testAccCheckBlockAccessValues(&blockAccess, expectedBlockAccessValues{
						manageActors: []api.ObjectActorAccess{
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
				Config: fixtureAccBlockAccess(baseCfg),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockAccessDestroy("prefect_block.test"),
				),
			},
		},
	})
}

// testAccCheckBlockAccessExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckBlockAccessExists(blockAccessResourceName string, blockAccess *api.BlockDocumentAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		blockID, err := testutils.GetResourceIDFromStateByAttribute(s, blockAccessResourceName, "block_id")
		if err != nil {
			return fmt.Errorf("error fetching block ID: %w", err)
		}

		workspaceID, err := testutils.GetResourceIDFromState(s, testutils.WorkspaceResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		blockClient, _ := c.BlockDocuments(uuid.Nil, workspaceID)

		fetchedBlockAccess, err := blockClient.GetAccess(context.Background(), blockID)
		if err != nil {
			return fmt.Errorf("error fetching block access: %w", err)
		}

		*blockAccess = *fetchedBlockAccess

		return nil
	}
}

type expectedBlockAccessValues struct {
	manageActors []api.ObjectActorAccess
	viewActors   []api.ObjectActorAccess
}

// testAccCheckBlockAccessValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckBlockAccessValues(fetchedBlockAccess *api.BlockDocumentAccess, expectedValues expectedBlockAccessValues) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		tests := map[string]struct {
			fetched  []api.ObjectActorAccess
			expected []api.ObjectActorAccess
		}{
			"manageActors": {fetchedBlockAccess.ManageActors, expectedValues.manageActors},
			"viewActors":   {fetchedBlockAccess.ViewActors, expectedValues.viewActors},
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

// testAccCheckBlockAccessDestroy is a Custom Check Function that
// verifies that the access control was reset to wildcard on deletion.
func testAccCheckBlockAccessDestroy(blockResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		workspaceID, err := testutils.GetResourceIDFromState(s, testutils.WorkspaceResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		blockID, err := testutils.GetResourceIDFromState(s, blockResourceName)
		if err != nil {
			return fmt.Errorf("error fetching block ID: %w", err)
		}

		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		blockClient, _ := c.BlockDocuments(uuid.Nil, workspaceID)

		fetchedBlockAccess, err := blockClient.GetAccess(context.Background(), blockID)
		if err != nil {
			return fmt.Errorf("error fetching block access: %w", err)
		}

		expectedActors := []api.ObjectActorAccess{
			{ID: "*", Name: "*", Type: api.AllAccessors},
		}

		checks := map[string][]api.ObjectActorAccess{
			"manage_actors": fetchedBlockAccess.ManageActors,
			"view_actors":   fetchedBlockAccess.ViewActors,
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
