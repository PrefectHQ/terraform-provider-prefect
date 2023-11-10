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
)

func fixtureAccWorkspaceCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	name = "%s"
	handle = "%s"
}
`, name, name)
}

func fixtureAccWorkspaceUpdate(name string, description string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	name = "%s"
	handle = "%s"
	description = "%s"
}`, name, name, description)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_workspace(t *testing.T) {
	resourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	emptyDescription := ""
	randomDescription := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspace api.Workspace

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the workspace resource
				Config: fixtureAccWorkspaceCreate(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName, &workspace),
					testAccCheckWorkspaceValues(&workspace, &api.Workspace{Name: randomName, Handle: randomName, Description: &emptyDescription}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "handle", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				// Check update of existing workspace resource
				Config: fixtureAccWorkspaceUpdate(randomName2, randomDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(resourceName, &workspace),
					testAccCheckWorkspaceValues(&workspace, &api.Workspace{Name: randomName2, Handle: randomName2, Description: &randomDescription}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "handle", randomName2),
					resource.TestCheckResourceAttr(resourceName, "description", randomDescription),
				),
			},
			// Import State checks - import by handle
			{
				ImportState:         true,
				ResourceName:        resourceName,
				ImportStateId:       randomName2,
				ImportStateIdPrefix: "handle/",
				ImportStateVerify:   true,
			},
			// Import State checks - import by ID (default)
			{
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWorkspaceExists(workspaceResourceName string, workspace *api.Workspace) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceResource, found := state.RootModule().Resources[workspaceResourceName]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", workspaceResourceName)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		workspacesClient, _ := c.Workspaces(uuid.Nil)
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		fetchedWorkspace, err := workspacesClient.Get(context.Background(), workspaceID)
		if err != nil {
			return fmt.Errorf("Error fetching workspace: %w", err)
		}
		if fetchedWorkspace == nil {
			return fmt.Errorf("Workspace not found for ID: %s", workspaceResource.Primary.ID)
		}

		*workspace = *fetchedWorkspace

		return nil
	}
}

func testAccCheckWorkspaceValues(fetchedWorkspace *api.Workspace, valuesToCheck *api.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if fetchedWorkspace.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected workspace name %s, got: %s", fetchedWorkspace.Name, valuesToCheck.Name)
		}
		if fetchedWorkspace.Handle != valuesToCheck.Handle {
			return fmt.Errorf("Expected workspace handle %s, got: %s", fetchedWorkspace.Handle, valuesToCheck.Handle)
		}
		if *fetchedWorkspace.Description != *valuesToCheck.Description {
			return fmt.Errorf("Expected workspace description %s, got: %s", *fetchedWorkspace.Description, *valuesToCheck.Description)
		}

		return nil
	}
}
