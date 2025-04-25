package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_workspace(t *testing.T) {
	// Workspace is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	ephemeralWorkspaceCreate := testutils.NewEphemeralWorkspace()
	ephemeralWorkspaceUpdate := testutils.NewEphemeralWorkspace()

	resourceName := testutils.WorkspaceResourceName

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workspace api.Workspace

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the workspace resource
				Config: ephemeralWorkspaceCreate.Resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(&workspace),
					testAccCheckWorkspaceValues(&workspace, &api.Workspace{Name: ephemeralWorkspaceCreate.Name, Handle: ephemeralWorkspaceCreate.Name, Description: &ephemeralWorkspaceCreate.Description}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", ephemeralWorkspaceCreate.Name),
					testutils.ExpectKnownValue(resourceName, "handle", ephemeralWorkspaceCreate.Name),
					testutils.ExpectKnownValue(resourceName, "description", ephemeralWorkspaceCreate.Description),
				},
			},
			{
				// Check update of existing workspace resource
				Config: ephemeralWorkspaceUpdate.Resource,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(&workspace),
					testAccCheckWorkspaceValues(&workspace, &api.Workspace{Name: ephemeralWorkspaceUpdate.Name, Handle: ephemeralWorkspaceUpdate.Name, Description: &ephemeralWorkspaceUpdate.Description}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", ephemeralWorkspaceUpdate.Name),
					testutils.ExpectKnownValue(resourceName, "handle", ephemeralWorkspaceUpdate.Name),
					testutils.ExpectKnownValue(resourceName, "description", ephemeralWorkspaceUpdate.Description),
				},
			},
			// Import State checks - import by handle
			{
				ImportState:         true,
				ResourceName:        resourceName,
				ImportStateId:       ephemeralWorkspaceUpdate.Name,
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

func testAccCheckWorkspaceExists(workspace *api.Workspace) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workspaceID, err := testutils.GetResourceIDFromState(state, testutils.WorkspaceResourceName)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		workspacesClient, _ := c.Workspaces(uuid.Nil)

		fetchedWorkspace, err := workspacesClient.Get(context.Background(), workspaceID)
		if err != nil {
			return fmt.Errorf("Error fetching workspace: %w", err)
		}
		if fetchedWorkspace == nil {
			return fmt.Errorf("Workspace not found for ID: %s", workspaceID)
		}

		*workspace = *fetchedWorkspace

		return nil
	}
}

func testAccCheckWorkspaceValues(fetchedWorkspace *api.Workspace, valuesToCheck *api.Workspace) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
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
