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

func fixtureAccWorkPoolCreate(name string, poolType string, paused bool) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	name = "%s"
	handle = "%s"
}
resource "prefect_work_pool" "test" {
	name = "%s"
	type = "%s"
	workspace_id = prefect_workspace.workspace.id
	paused = %t
	depends_on = [prefect_workspace.workspace]
}
`, name, name, name, poolType, paused)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_work_pool(t *testing.T) {
	workPoolResourceName := "prefect_work_pool.test"
	workspaceResourceName := "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	poolType := "kubernetes"
	poolType2 := "ecs"

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var workPool api.WorkPool

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the work pool resource
				Config: fixtureAccWorkPoolCreate(randomName, poolType, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkPoolExists(workPoolResourceName, workspaceResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName, Type: poolType, IsPaused: true}),
					resource.TestCheckResourceAttr(workPoolResourceName, "name", randomName),
					resource.TestCheckResourceAttr(workPoolResourceName, "type", poolType),
					resource.TestCheckResourceAttr(workPoolResourceName, "paused", "true"),
				),
			},
			{
				// Check that changing the paused state will update the resource in place
				Config: fixtureAccWorkPoolCreate(randomName, poolType, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDAreEqual(workPoolResourceName, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName, workspaceResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName, Type: poolType, IsPaused: false}),
					resource.TestCheckResourceAttr(workPoolResourceName, "name", randomName),
					resource.TestCheckResourceAttr(workPoolResourceName, "type", poolType),
					resource.TestCheckResourceAttr(workPoolResourceName, "paused", "false"),
				),
			},
			{
				// Check that changing the name will re-create the resource
				Config: fixtureAccWorkPoolCreate(randomName2, poolType, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDsNotEqual(workPoolResourceName, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName, workspaceResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName2, Type: poolType, IsPaused: false}),
				),
			},
			{
				// Check that changing the poolType will re-create the resource
				Config: fixtureAccWorkPoolCreate(randomName2, poolType2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDsNotEqual(workPoolResourceName, &workPool),
					testAccCheckWorkPoolExists(workPoolResourceName, workspaceResourceName, &workPool),
					testAccCheckWorkPoolValues(&workPool, &api.WorkPool{Name: randomName2, Type: poolType2, IsPaused: false}),
				),
			},
			// Import State checks - import by workspace_id,name (dynamic)
			{
				ImportState:       true,
				ResourceName:      workPoolResourceName,
				ImportStateIdFunc: getWorkPoolImportStateID(workPoolResourceName, workspaceResourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWorkPoolExists(workPoolResourceName string, workspaceResourceName string, workPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolResource, exists := state.RootModule().Resources[workPoolResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workPoolResourceName)
		}

		workspaceResource, exists := state.RootModule().Resources[workspaceResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		workPoolsClient, _ := c.WorkPools(uuid.Nil, workspaceID)

		workPoolName := workPoolResource.Primary.Attributes["name"]

		fetchedWorkPool, err := workPoolsClient.Get(context.Background(), workPoolName)
		if err != nil {
			return fmt.Errorf("Error fetching work pool: %w", err)
		}
		if fetchedWorkPool == nil {
			return fmt.Errorf("Work Pool not found for name: %s", workPoolName)
		}

		*workPool = *fetchedWorkPool

		return nil
	}
}

func testAccCheckWorkPoolValues(fetchedWorkPool *api.WorkPool, valuesToCheck *api.WorkPool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedWorkPool.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected work pool name to be %s, got %s", valuesToCheck.Name, fetchedWorkPool.Name)
		}

		if fetchedWorkPool.Type != valuesToCheck.Type {
			return fmt.Errorf("Expected work pool type to be %s, got %s", valuesToCheck.Type, fetchedWorkPool.Type)
		}

		if fetchedWorkPool.IsPaused != valuesToCheck.IsPaused {
			return fmt.Errorf("Expected work pool paused to be %t, got %t", valuesToCheck.IsPaused, fetchedWorkPool.IsPaused)
		}

		return nil
	}
}

func testAccCheckIDAreEqual(resourceName string, fetchedWorkPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkPool.ID.String()

		if workPoolResource.Primary.ID != id {
			return fmt.Errorf("Expected %s and %s to be equal", workPoolResource.Primary.ID, id)
		}

		return nil
	}
}

func testAccCheckIDsNotEqual(resourceName string, fetchedWorkPool *api.WorkPool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		workPoolResource, exists := state.RootModule().Resources[resourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		id := fetchedWorkPool.ID.String()

		if workPoolResource.Primary.ID == id {
			return fmt.Errorf("Expected %s and %s to be different", workPoolResource.Primary.ID, id)
		}

		return nil
	}
}

func getWorkPoolImportStateID(workPoolResourceName string, workspaceResourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceResource, exists := state.RootModule().Resources[workspaceResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		workPoolResource, exists := state.RootModule().Resources[workPoolResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workPoolResourceName)
		}
		workPoolName := workPoolResource.Primary.Attributes["name"]

		return fmt.Sprintf("%s,%s", workspaceID, workPoolName), nil
	}
}
