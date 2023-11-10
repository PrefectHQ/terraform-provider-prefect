package resources_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWorkspaceRoleResource(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace_role" "role" {
	name = "%s"
	description = "%s description"
	scopes = ["see_blocks", "see_artifacts"]
}`, name, name)
}

func fixtureAccWorkspaceRoleReesourceUpdated(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace_role" "role" {
	name = "%s"
	description = "description for %s"
	scopes = ["see_workers", "see_variables", "see_work_queues"]
}`, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_workspace_role(t *testing.T) {
	resourceName := "prefect_workspace_role.role"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspaceRole api.WorkspaceRole

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the workspace role resource
				Config: fixtureAccWorkspaceRoleResource(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceRoleExists(resourceName, &workspaceRole),
					testAccCheckWorkspaceRoleValues(&workspaceRole, &api.WorkspaceRole{Name: randomName, Scopes: []string{"see_artifacts", "see_blocks"}}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("%s description", randomName)),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scopes.0", "see_blocks"),
					resource.TestCheckResourceAttr(resourceName, "scopes.1", "see_artifacts"),
				),
			},
			{
				// Check updates for the workspace role resource
				Config: fixtureAccWorkspaceRoleReesourceUpdated(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceRoleExists(resourceName, &workspaceRole),
					testAccCheckWorkspaceRoleValues(&workspaceRole, &api.WorkspaceRole{Name: randomName, Scopes: []string{"see_workers", "see_work_queues", "see_variables"}}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("description for %s", randomName)),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "scopes.0", "see_workers"),
					resource.TestCheckResourceAttr(resourceName, "scopes.1", "see_variables"),
					resource.TestCheckResourceAttr(resourceName, "scopes.2", "see_work_queues"),
				),
			},
			// Import State checks - import by ID (default)
			{
				ImportState:             true,
				ResourceName:            resourceName,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scopes"},
			},
		},
	})
}

func testAccCheckWorkspaceRoleExists(roleResourceName string, role *api.WorkspaceRole) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceRoleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", roleResourceName)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		workspaceRolesClient, _ := c.WorkspaceRoles(uuid.Nil)
		resourceID, _ := uuid.Parse(workspaceRoleResource.Primary.ID)

		fetchedWorkspaceRole, err := workspaceRolesClient.Get(context.Background(), resourceID)
		if err != nil {
			return fmt.Errorf("Error fetching Workspace Role: %w", err)
		}
		if fetchedWorkspaceRole == nil {
			return fmt.Errorf("Workspace Role not found for ID: %s", workspaceRoleResource.Primary.ID)
		}

		*role = *fetchedWorkspaceRole

		return nil
	}
}

func testAccCheckWorkspaceRoleValues(fetchedRole *api.WorkspaceRole, valuesToCheck *api.WorkspaceRole) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if fetchedRole.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected Workspace Role name %s, got: %s", fetchedRole.Name, valuesToCheck.Name)
		}

		sort.StringSlice(fetchedRole.Scopes).Sort()
		sort.StringSlice(valuesToCheck.Scopes).Sort()

		if !reflect.DeepEqual(fetchedRole.Scopes, valuesToCheck.Scopes) {
			return fmt.Errorf("Expected Workspace Role scopes %v, got: %v", fetchedRole.Scopes, valuesToCheck.Scopes)
		}

		return nil
	}
}
