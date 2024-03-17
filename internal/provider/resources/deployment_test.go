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

func fixtureAccDeploymentCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_deployment" "deployment" {
	name = "%s"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
}
`, name)
}

func fixtureAccDeploymentUpdate(name string, description string) string {
	return fmt.Sprintf(`
resource "prefect_deployment" "deployment" {
	name = "%s"
	handle = "%s"
	description = "%s"
}`, name, name, description)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	resourceName := "prefect_deployment.deployment"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// emptyDescription := ""
	randomDescription := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var deployment api.Deployment

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccDeploymentCreate(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					testAccCheckDeploymentValues(&deployment, &api.Deployment{
						Name: randomName,
						// Handle: randomName,
						// Description: &emptyDescription,
					}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					// resource.TestCheckResourceAttr(resourceName, "handle", randomName),
					// resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				// Check update of existing deployment resource
				Config: fixtureAccDeploymentUpdate(randomName2, randomDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, &deployment),
					testAccCheckDeploymentValues(&deployment, &api.Deployment{
						Name: randomName2,
						// Handle: randomName2,
						// Description: &randomDescription,
					}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					// resource.TestCheckResourceAttr(resourceName, "handle", randomName2),
					// resource.TestCheckResourceAttr(resourceName, "description", randomDescription),
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

func testAccCheckDeploymentExists(deploymentResourceName string, deployment *api.Deployment) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		deploymentResource, found := state.RootModule().Resources[deploymentResourceName]
		if !found {
			return fmt.Errorf("Resource not found in state: %s", deploymentResourceName)
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		deploymentsClient, _ := c.Deployments(uuid.Nil, uuid.Nil)
		deploymentID, _ := uuid.Parse(deploymentResource.Primary.ID)

		fetchedDeployment, err := deploymentsClient.Get(context.Background(), deploymentID)
		if err != nil {
			return fmt.Errorf("Error fetching deployment: %w", err)
		}
		if fetchedDeployment == nil {
			return fmt.Errorf("Deployment not found for ID: %s", deploymentResource.Primary.ID)
		}

		*deployment = *fetchedDeployment

		return nil
	}
}

func testAccCheckDeploymentValues(fetchedDeployment *api.Deployment, valuesToCheck *api.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if fetchedDeployment.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected workspace name %s, got: %s", fetchedDeployment.Name, valuesToCheck.Name)
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
