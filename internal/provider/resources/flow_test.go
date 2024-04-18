package resources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccFlowCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}

resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = prefect_workspace.workspace.id
	tags = ["test"]
}
`, name, name, name)
}

// func fixtureAccFlowUpdate(name string) string {
// 	return fmt.Sprintf(`
// resource "prefect_flow" "flow" {
// 	name = "%s"
// 	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
// 	tags = ["test1"]
// }`, name)
// }

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_flow(t *testing.T) {
	resourceName := "prefect_flow.flow"
	workspaceResourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// emptyDescription := ""
	// randomDescription := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccFlowCreate(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
				),
			},
			// Import State checks - import by ID (default)
			{
				ImportState:       true,
				ImportStateIdFunc: getFlowImportStateID(resourceName, workspaceResourceName),
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}

func getFlowImportStateID(flowName string, workspaceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspace, exists := state.RootModule().Resources[workspaceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workspaceName)
		}
		workspaceID, _ := uuid.Parse(workspace.Primary.ID)

		flow, exists := state.RootModule().Resources[flowName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", flowName)
		}

		return fmt.Sprintf("%s,%s", workspaceID, flow.Primary.Attributes["id"]), nil
	}
}
