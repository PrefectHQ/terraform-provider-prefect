package helpers

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func GetResourceWorkspaceImportStateID(resourceName string, workspaceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspace, exists := state.RootModule().Resources[workspaceName]
		if !exists {
			return "", fmt.Errorf("resource not found in state: %s", workspaceName)
		}
		workspaceID, _ := uuid.Parse(workspace.Primary.ID)

		resource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return "", fmt.Errorf("resource not found in state: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", workspaceID, resource.Primary.Attributes["id"]), nil
	}
}
