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

func fixtureAccFlowCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	tags = ["test"]
}
`, name)
}

func fixtureAccFlowUpdate(name string) string {
	return fmt.Sprintf(`
resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	tags = ["test1"]
}`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_flow(t *testing.T) {
	resourceName := "prefect_flow.flow"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// const workspaceResourceName = "data.prefect_workspace.evergreen"
	const workspaceResourceName = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	// randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// emptyDescription := ""
	// randomDescription := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var flow api.Flow

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccFlowCreate(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFlowExists(resourceName, workspaceResourceName, &flow),
					testAccCheckFlowValues(&flow, &api.Flow{
						Name: randomName,
					}),
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

func testAccCheckFlowExists(flowResourceName string, workspaceResouceName string, flow *api.Flow) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, found := state.RootModule().Resources[flowResourceName]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", flowResourceName)
		}

		// workspaceResource, found := state.RootModule().Resources[workspaceResouceName]
		// if !found {
		// 	return fmt.Errorf("Resource not found in state: %s", workspaceResouceName)
		// }
		// workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)
		workspaceID, _ := uuid.Parse(workspaceResouceName)

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		flowsClient, _ := c.Flows(uuid.Nil, workspaceID)
		flowID, _ := uuid.Parse(flowResource.Primary.ID)

		fetchedFlow, err := flowsClient.Get(context.Background(), flowID)
		if err != nil {
			return fmt.Errorf("Error fetching deployment: %w", err)
		}
		if fetchedFlow == nil {
			return fmt.Errorf("Deployment not found for ID: %s", flowResource.Primary.ID)
		}

		*flow = *fetchedFlow

		return nil
	}
}

func testAccCheckFlowValues(fetchedFlow *api.Flow, valuesToCheck *api.Flow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if fetchedFlow.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected flow name %s, got: %s", fetchedFlow.Name, valuesToCheck.Name)
		}
		// if fetchedDeployment.Handle != valuesToCheck.Handle {
		// 	return fmt.Errorf("Expected workspace handle %s, got: %s", fetchedDeployment.Handle, valuesToCheck.Handle)
		// }
		// if *fetchedDeployment.Description != *valuesToCheck.Description {
		// 	return fmt.Errorf("Expected workspace description %s, got: %s", *fetchedDeployment.Description, *valuesToCheck.Description)
		// }

		return nil
	}
}
