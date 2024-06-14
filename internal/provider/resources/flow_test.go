package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccFlowCreate(name string, tag string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	handle = "%s"
	name = "%s"
}

resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = prefect_workspace.workspace.id
	tags = ["%s"]
}
`, name, name, name, tag)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_flow(t *testing.T) {
	resourceName := "prefect_flow.flow"
	workspaceResourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccFlowCreate(randomName, "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test1"),
				),
			},
			{
				// Check updating the resource
				Config: fixtureAccFlowCreate(randomName, "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test2"),
				),
			},
			// Import State checks - import by ID (default)
			{
				ImportState:       true,
				ImportStateIdFunc: helpers.GetResourceWorkspaceImportStateID(resourceName, workspaceResourceName),
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}
