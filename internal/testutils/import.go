package testutils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// getResourceWorkspaceImportStateID is a helper function that standardizes the format used
// for importing a resource in the format "<identifier>,<workspace_id>".
func getResourceWorkspaceImportStateID(resourceName, identifier string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspace, exists := state.RootModule().Resources[WorkspaceResourceName]
		if !exists {
			return "", fmt.Errorf("resource not found in state: %s", WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspace.Primary.ID)

		fetchedResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return "", fmt.Errorf("resource not found in state: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", fetchedResource.Primary.Attributes[identifier], workspaceID), nil
	}
}

// GetResourceWorkspaceImportStateID is a helper function that returns a resource.ImportStateIdFunc
// that can be used to import a resource by its ID and workspace ID.
func GetResourceWorkspaceImportStateID(resourceName string) resource.ImportStateIdFunc {
	return getResourceWorkspaceImportStateID(resourceName, "id")
}

// GetResourceWorkspaceImportStateIDByName is a helper function that returns a resource.ImportStateIdFunc
// that can be used to import a resource by its name and workspace ID.
func GetResourceWorkspaceImportStateIDByName(resourceName string) resource.ImportStateIdFunc {
	return getResourceWorkspaceImportStateID(resourceName, "name")
}
