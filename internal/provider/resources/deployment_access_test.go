package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentAccessConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
}

func fixtureAccDeploymentAccess(cfg deploymentAccessConfig) string {
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

resource "prefect_deployment" "test" {
	name = "my-deployment"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_access" "test" {
	workspace_id = {{.WorkspaceResourceName}}.id
	deployment_id = prefect_deployment.test.id

	manage_actor_ids = [prefect_service_account.test.actor_id]
	run_actor_ids = [prefect_service_account.test.actor_id]
	view_actor_ids = [prefect_service_account.test.actor_id]

	manage_team_ids = [data.prefect_team.test.id]
	run_team_ids = [data.prefect_team.test.id]
	view_team_ids = [data.prefect_team.test.id]
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_access(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	serviceAccountName := "my-service-account"
	teamName := "my-team"

	cfgSet := deploymentAccessConfig{
		WorkspaceResource:     workspace.Resource,
		WorkspaceResourceName: testutils.WorkspaceResourceName,
	}

	var deployment api.Deployment
	var deploymentAccess api.DeploymentAccessControl

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentAccess(cfgSet),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists("prefect_deployment.test", cfgSet.WorkspaceResourceName, &deployment),
					testAccCheckDeploymentAccessExists("prefect_deployment_access.test", &deploymentAccess),
					testAccCheckDeploymentAccessValues(&deploymentAccess, expectedDeploymentAccessValues{
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

// testAccCheckDeploymentAccessExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckDeploymentAccessExists(deploymentAccessResourceName string, deploymentAccess *api.DeploymentAccessControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the deployment access resource we just created from the state
		deploymentAccessResource, exists := s.RootModule().Resources[deploymentAccessResourceName]
		if !exists {
			return fmt.Errorf("deployment access resource not found: %s", deploymentAccessResourceName)
		}
		deploymentAccessID, _ := uuid.Parse(deploymentAccessResource.Primary.Attributes["deployment_id"])

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("workspace resource not found: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		deploymentAccessClient, _ := c.DeploymentAccess(uuid.Nil, workspaceID)

		fetchedDeploymentAccess, err := deploymentAccessClient.Read(context.Background(), deploymentAccessID)
		if err != nil {
			return fmt.Errorf("error fetching deployment access: %w", err)
		}

		// Assign the fetched deployment to the passed pointer
		// so we can use it in the next test assertion
		*deploymentAccess = *fetchedDeploymentAccess

		return nil
	}
}

type expectedDeploymentAccessValues struct {
	manageActors []api.ObjectActorAccess
	runActors    []api.ObjectActorAccess
	viewActors   []api.ObjectActorAccess
}

// testAccCheckDeploymentValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckDeploymentAccessValues(fetchedDeploymentAccess *api.DeploymentAccessControl, expectedValues expectedDeploymentAccessValues) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		tests := map[string]struct {
			fetched  []api.ObjectActorAccess
			expected []api.ObjectActorAccess
		}{
			"manageActors": {fetchedDeploymentAccess.ManageActors, expectedValues.manageActors},
			"runActors":    {fetchedDeploymentAccess.RunActors, expectedValues.runActors},
			"viewActors":   {fetchedDeploymentAccess.ViewActors, expectedValues.viewActors},
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

func actorFound(fetched []api.ObjectActorAccess, expected []api.ObjectActorAccess) error {
	if len(fetched) != len(expected) {
		return fmt.Errorf("got %d actors, expected %d", len(fetched), len(expected))
	}

	for i := range expected {
		found := false
		for j := range fetched {
			if fetched[j].Name == expected[i].Name && fetched[j].Type == expected[i].Type {
				found = true

				break
			}
		}

		if !found {
			return fmt.Errorf("actor %s of type %s not found", expected[i].Name, expected[i].Type)
		}
	}

	return nil
}
