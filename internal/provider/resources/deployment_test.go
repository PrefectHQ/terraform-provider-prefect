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
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccDeploymentCreate(name, description string) string {
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

resource "prefect_deployment" "deployment" {
	name = "%s"
	description = "%s"
  enforce_parameter_schema = false
	entrypoint = "hello_world.py:hello_world"
	flow_id = prefect_flow.flow.id
  manifest_path            = "./bar/foo"
  path                     = "./foo/bar"
  paused                   = false
	tags = ["test"]
  version                  = "v1.1.1"
  work_pool_name           = "test-pool"
  work_queue_name          = "default"
	workspace_id = prefect_workspace.workspace.id
}
`, name, name, name, name, description)
}

func fixtureAccDeploymentUpdate(name string, description string) string {
	return fmt.Sprintf(`
resource "prefect_deployment" "deployment" {
	name = "%s"
	description = "%s"
}`, name, description)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	resourceName := "prefect_deployment.deployment"
	workspaceResourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	description := "My deployment description"
	description2 := "My deployment description v2"

	var deployment api.Deployment

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccDeploymentCreate(randomName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				// Check update of existing deployment resource
				Config: fixtureAccDeploymentUpdate(randomName2, description2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(resourceName, workspaceResourceName, &deployment),
					testAccCheckDeploymentValues(&deployment, expectedDeploymentValues{
						name:        randomName2,
						description: description2,
					}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "description", description2),
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

// testAccCheckDeploymentExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckDeploymentExists(deploymentResourceName string, workspaceResourceName string, deployment *api.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the deployment resource we just created from the state
		deploymentResource, exists := s.RootModule().Resources[deploymentResourceName]
		if !exists {
			return fmt.Errorf("deployment resource not found: %s", deploymentResourceName)
		}
		deploymentID, _ := uuid.Parse(deploymentResource.Primary.ID)

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[workspaceResourceName]
		if !exists {
			return fmt.Errorf("workspace resource not found: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		deploymentsClient, _ := c.Deployments(uuid.Nil, workspaceID)

		fetchedDeployment, err := deploymentsClient.Get(context.Background(), deploymentID)
		if err != nil {
			return fmt.Errorf("error fetching deployment: %w", err)
		}

		// Assign the fetched deployment to the passed pointer
		// so we can use it in the next test assertion
		*deployment = *fetchedDeployment

		return nil
	}
}

type expectedDeploymentValues struct {
	name        string
	description string
}

// testAccCheckDeploymentValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckDeploymentValues(fetchedDeployment *api.Deployment, expectedValues expectedDeploymentValues) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedDeployment.Name != expectedValues.name {
			return fmt.Errorf("Expected deployment name to be %s, got %s", expectedValues.name, fetchedDeployment.Name)
		}
		if fetchedDeployment.Description != expectedValues.description {
			return fmt.Errorf("Expected deployment description to be %s, got %s", expectedValues.description, fetchedDeployment.Description)
		}

		return nil
	}
}
