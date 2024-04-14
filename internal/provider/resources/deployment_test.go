package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccDeploymentCreate(name string) string {
	return fmt.Sprintf(`
resource "prefect_flow" "flow" {
	name = "%s"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	tags = ["test"]
}

resource "prefect_deployment" "deployment" {
	name = "%s"
	description = "string"
	workspace_id = "7e6f15bf-487a-4811-83ef-f074ec6c5484"
	flow_id = prefect_flow.flow.id
	entrypoint = "hello_world.py:hello_world"
	tags = ["test"]

	// "schedules": [
	// 	{
	// 		"active": true,
	// 		"schedule": {
	// 			"interval": 0,
	// 			"anchor_date": "2019-08-24T14:15:22Z",
	// 			"timezone": "America/New_York"
	// 		}
	// 	}
	// ],
	// parameters: {
	// 	'goodbye': True
	// },
	// parameter_openapi_schema = {
	// 	'title': 'Parameters',
	// 	'type': 'object',
	// 	'properties': {
	// 		'name': {...},
	// 		'goodbye': {...}
	// 	}
	// }

	// is_schedule_active = true
	// paused = false
	// 
	// enforce_parameter_schema = false
	// "parameter_openapi_schema": { },
	// "pull_steps": [
	// { }
	// ],
	
	// manifest_path = "string"
	// work_queue_name = "string"
	// work_pool_name = "my-work-pool"
	// storage_document_id = "e0212ac4-7ec3-401b-b1e6-2a4627d8e7ec"
	// infrastructure_document_id = "ce9a08a7-d77b-4b3f-826a-53820cfe01fa"
	// schedule = {
	// 	interval = 0
	// 	anchor_date = "2019-08-24T14:15:22Z"
	// 	timezone = "America/New_York"
	// },
	// path = "string"
	// version = "string"
	// infra_overrides = { }
}
`, name, name)
}

// func fixtureAccDeploymentUpdate(name string, description string) string {
// 	return fmt.Sprintf(`
// resource "prefect_deployment" "deployment" {
// 	name = "%s"
// 	handle = "%s"
// 	description = "%s"
// }`, name, name, description)
// }

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	resourceName := "prefect_deployment.deployment"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	// emptyDescription := ""
	// randomDescription := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	// var deployment api.Deployment

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccDeploymentCreate(randomName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					// resource.TestCheckResourceAttr(resourceName, "handle", randomName),
					// resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			// {
			// 	// Check update of existing deployment resource
			// 	Config: fixtureAccDeploymentUpdate(randomName2, randomDescription),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckDeploymentExists(resourceName, &deployment),
			// 		testAccCheckDeploymentValues(&deployment, &api.Deployment{
			// 			Name: randomName2,
			// 			// Handle: randomName2,
			// 			// Description: &randomDescription,
			// 		}),
			// 		resource.TestCheckResourceAttr(resourceName, "name", randomName2),
			// 		// resource.TestCheckResourceAttr(resourceName, "handle", randomName2),
			// 		// resource.TestCheckResourceAttr(resourceName, "description", randomDescription),
			// 	),
			// },
			// // Import State checks - import by handle
			// {
			// 	ImportState:         true,
			// 	ResourceName:        resourceName,
			// 	ImportStateId:       randomName2,
			// 	ImportStateIdPrefix: "handle/",
			// 	ImportStateVerify:   true,
			// },
			// // Import State checks - import by ID (default)
			// {
			// 	ImportState:       true,
			// 	ResourceName:      resourceName,
			// 	ImportStateVerify: true,
			// },
		},
	})
}
