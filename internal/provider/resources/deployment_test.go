package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentConfig struct {
	workspace             string
	workspaceName         string
	workspaceResourceName string

	deploymentName         string
	deploymentResourceName string

	description string

	flowName string
}

func fixtureAccDeployment(cfg deploymentConfig) string {
	return fmt.Sprintf(`
%s

resource "prefect_flow" "%s" {
	name = "%s"
	tags = ["test"]

	workspace_id = prefect_workspace.%s.id
	depends_on = [prefect_workspace.%s]
}

resource "prefect_deployment" "%s" {
	name = "%s"
	description = "%s"
  enforce_parameter_schema = false
	entrypoint = "hello_world.py:hello_world"
	flow_id = prefect_flow.%s.id
  manifest_path            = "./bar/foo"
  path                     = "./foo/bar"
  paused                   = false
	tags = ["test"]
  version                  = "v1.1.1"

	workspace_id = prefect_workspace.%s.id
	depends_on = [prefect_workspace.%s, prefect_flow.%s]
}
`, cfg.workspace,
		cfg.flowName,
		cfg.flowName,
		cfg.workspaceName,
		cfg.workspaceName,
		cfg.deploymentName,
		cfg.deploymentName,
		cfg.description,
		cfg.flowName,
		cfg.workspaceName,
		cfg.workspaceName,
		cfg.flowName,
	)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	workspace, workspaceName := testutils.NewEphemeralWorkspace()

	cfgCreate := deploymentConfig{
		workspace:      workspace,
		workspaceName:  workspaceName,
		deploymentName: testutils.NewRandomPrefixedString(),
		flowName:       testutils.NewRandomPrefixedString(),
		description:    "My deployment description",
	}

	cfgCreate.deploymentResourceName = fmt.Sprintf("prefect_deployment.%s", cfgCreate.deploymentName)
	cfgCreate.workspaceResourceName = fmt.Sprintf("prefect_workspace.%s", cfgCreate.workspaceName)

	cfgUpdate := cfgCreate
	cfgUpdate.description = "My deployment description v2"

	// var deployment api.Deployment

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccDeployment(cfgCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(cfgCreate.deploymentResourceName, "name", cfgCreate.deploymentName),
					resource.TestCheckResourceAttr(cfgCreate.deploymentResourceName, "description", cfgCreate.description),
				),
			},
			// {
			// 	// Check update of existing deployment resource
			// 	Config: fixtureAccDeployment(cfgUpdate),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckDeploymentExists(cfgUpdate.deploymentResourceName, cfgUpdate.workspaceResourceName, &deployment),
			// 		testAccCheckDeploymentValues(&deployment, expectedDeploymentValues{
			// 			name:        cfgUpdate.deploymentName,
			// 			description: cfgUpdate.description,
			// 		}),
			// 		resource.TestCheckResourceAttr(cfgUpdate.deploymentResourceName, "name", cfgUpdate.deploymentName),
			// 		resource.TestCheckResourceAttr(cfgUpdate.deploymentResourceName, "description", cfgUpdate.description),
			// 	),
			// },
			// Import State checks - import by ID (default)
			{
				ImportState:       true,
				ImportStateIdFunc: helpers.GetResourceWorkspaceImportStateID(cfgCreate.deploymentResourceName, cfgCreate.workspaceResourceName),
				ResourceName:      cfgCreate.deploymentResourceName,
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
