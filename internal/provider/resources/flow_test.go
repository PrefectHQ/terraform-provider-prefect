package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccFlowCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	tags = ["test"]
}
`, name)
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
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// const workspaceResourceName = "data.prefect_workspace.evergreen"
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
			// {
			// 	ImportState:       true,
			// 	ImportStateId:      workspaceResourceName + "," + flow.ID.String(),
			// 	ResourceName: 		resourceName,
			// 	ImportStateVerify: true,
			// },
		},
	})
}
