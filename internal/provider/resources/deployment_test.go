package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentConfig struct {
	Workspace string

	DeploymentName         string
	DeploymentResourceName string

	Description            string
	EnforceParameterSchema bool
	Entrypoint             string
	JobVariables           string
	ManifestPath           string
	Parameters             string
	Path                   string
	Paused                 bool
	Tags                   []string
	Version                string
	WorkPoolName           string
	WorkQueueName          string
	ParameterOpenAPISchema string

	FlowName string

	StorageDocumentName string
}

func fixtureAccDeployment(cfg deploymentConfig) string {
	tmpl := `
{{.Workspace}}

resource "prefect_block" "test_gh_repository" {
	name = "{{.StorageDocumentName}}"
	type_slug = "github-repository"

	data = jsonencode({
		"repository_url": "https://github.com/foo/bar",
		"reference": "main"
	})

	workspace_id = prefect_workspace.test.id
}

resource "prefect_work_pool" "{{.WorkPoolName}}" {
  name         = "{{.WorkPoolName}}"
  type         = "kubernetes"
  paused       = false
  workspace_id = prefect_workspace.test.id
}

resource "prefect_flow" "{{.FlowName}}" {
	name = "{{.FlowName}}"
	tags = [{{range .Tags}}"{{.}}", {{end}}]

	workspace_id = prefect_workspace.test.id
}

resource "prefect_deployment" "{{.DeploymentName}}" {
	name = "{{.DeploymentName}}"
	description = "{{.Description}}"
	enforce_parameter_schema = {{.EnforceParameterSchema}}
	entrypoint = "{{.Entrypoint}}"
	flow_id = prefect_flow.{{.FlowName}}.id
	job_variables = jsonencode(
		{{.JobVariables}}
	)
	manifest_path = "{{.ManifestPath}}"
	parameters = jsonencode({
		"some-parameter": "{{.Parameters}}"
	})
	path = "{{.Path}}"
	paused = {{.Paused}}
	tags = [{{range .Tags}}"{{.}}", {{end}}]
	version = "{{.Version}}"
	work_pool_name = "{{.WorkPoolName}}"
	work_queue_name = "{{.WorkQueueName}}"
	parameter_openapi_schema = jsonencode({{.ParameterOpenAPISchema}})
	storage_document_id = prefect_block.test_gh_repository.id

	workspace_id = prefect_workspace.test.id
	depends_on = [prefect_flow.{{.FlowName}}]
}
`
	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	deploymentName := testutils.NewRandomPrefixedString()
	flowName := testutils.NewRandomPrefixedString()

	parameterOpenAPISchema := `{"type": "object", "properties": {"some-parameter": {"type": "string"}}}`
	var parameterOpenAPISchemaMap map[string]interface{}
	_ = json.Unmarshal([]byte(parameterOpenAPISchema), &parameterOpenAPISchemaMap)

	cfgCreate := deploymentConfig{
		DeploymentName:         deploymentName,
		FlowName:               flowName,
		DeploymentResourceName: fmt.Sprintf("prefect_deployment.%s", deploymentName),
		Workspace:              workspace.Resource,

		Description:            "My deployment description",
		EnforceParameterSchema: false,
		Entrypoint:             "hello_world.py:hello_world",
		JobVariables:           `{"env":{"some-key":"some-value"}}`,
		ManifestPath:           "some-manifest-path",
		Parameters:             "some-value1",
		Path:                   "some-path",
		Paused:                 false,
		Tags:                   []string{"test1", "test2"},
		Version:                "v1.1.1",
		WorkPoolName:           "some-pool",
		WorkQueueName:          "default",
		ParameterOpenAPISchema: parameterOpenAPISchema,
		StorageDocumentName:    testutils.NewRandomPrefixedString(),
	}

	_ = deploymentConfig{
		// Keep some values from cfgCreate so we refer to the same resources for the update.
		DeploymentName:         cfgCreate.DeploymentName,
		FlowName:               cfgCreate.FlowName,
		DeploymentResourceName: cfgCreate.DeploymentResourceName,
		Workspace:              cfgCreate.Workspace,
		WorkPoolName:           cfgCreate.WorkPoolName,

		// Configure new values to test the update.
		Description:   "My deployment description v2",
		Entrypoint:    "hello_world.py:hello_world2",
		JobVariables:  `{"env":{"some-key":"some-value2"}}`,
		ManifestPath:  "some-manifest-path2",
		Parameters:    "some-value2",
		Path:          "some-path2",
		Paused:        true,
		Version:       "v1.1.2",
		WorkQueueName: "default",

		// Enforcing parameter schema  returns the following error:
		//
		//   Could not update deployment, unexpected error: status code 409 Conflict,
		//   error={"detail":"Error updating deployment: Cannot update parameters because
		//   parameter schema enforcement is enabledand the deployment does not have a
		//   valid parameter schema."}
		//
		// Will avoid testing this for now until a schema is configurable in the provider.
		//
		// EnforceParameterSchema: true
		EnforceParameterSchema: cfgCreate.EnforceParameterSchema,

		// Changing the tags results in a "404 Deployment not found" error.
		// Will avoid testing this until a solution is found.
		//
		// Tags: []string{"test1", "test3"}
		Tags: cfgCreate.Tags,

		// ParameterOpenAPISchema is not settable via the Update method.
		ParameterOpenAPISchema: cfgCreate.ParameterOpenAPISchema,

		StorageDocumentName: cfgCreate.StorageDocumentName,
	}

	var deployment api.Deployment

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the deployment resource
				Config: fixtureAccDeployment(cfgCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(cfgCreate.DeploymentResourceName, &deployment),
					testAccCheckDeploymentValues(&deployment, expectedDeploymentValues{
						name:                   cfgCreate.DeploymentName,
						description:            cfgCreate.Description,
						parameterOpenapiSchema: parameterOpenAPISchemaMap,
					}),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "enforce_parameter_schema", strconv.FormatBool(cfgCreate.EnforceParameterSchema)),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "entrypoint", cfgCreate.Entrypoint),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "job_variables", cfgCreate.JobVariables),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "manifest_path", cfgCreate.ManifestPath),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "parameters", `{"some-parameter":"some-value1"}`),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "path", cfgCreate.Path),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "paused", strconv.FormatBool(cfgCreate.Paused)),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "tags.0", cfgCreate.Tags[0]),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "tags.1", cfgCreate.Tags[1]),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "version", cfgCreate.Version),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "work_pool_name", cfgCreate.WorkPoolName),
					resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "work_queue_name", cfgCreate.WorkQueueName),
					resource.TestCheckResourceAttrSet(cfgCreate.DeploymentResourceName, "storage_document_id"),
				),
			},
			// {
			// 	// Check update of existing deployment resource
			// 	Config: fixtureAccDeployment(cfgUpdate),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckDeploymentExists(cfgUpdate.DeploymentResourceName, &deployment),
			// 		testAccCheckDeploymentValues(&deployment, expectedDeploymentValues{
			// 			name:                   cfgUpdate.DeploymentName,
			// 			description:            cfgUpdate.Description,
			// 			parameterOpenapiSchema: parameterOpenAPISchemaMap,
			// 		}),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "enforce_parameter_schema", strconv.FormatBool(cfgUpdate.EnforceParameterSchema)),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "entrypoint", cfgUpdate.Entrypoint),
			// 		resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "job_variables", cfgUpdate.JobVariables),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "manifest_path", cfgUpdate.ManifestPath),
			// 		resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "parameters", `{"some-parameter":"some-value2"}`),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "path", cfgUpdate.Path),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "paused", strconv.FormatBool(cfgUpdate.Paused)),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "tags.#", "2"),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "tags.0", cfgUpdate.Tags[0]),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "tags.1", cfgUpdate.Tags[1]),
			// 		resource.TestCheckResourceAttr(cfgUpdate.DeploymentResourceName, "version", cfgUpdate.Version),
			// 		resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "work_pool_name", cfgUpdate.WorkPoolName),
			// 		resource.TestCheckResourceAttr(cfgCreate.DeploymentResourceName, "work_queue_name", cfgUpdate.WorkQueueName),
			// 		resource.TestCheckResourceAttrSet(cfgCreate.DeploymentResourceName, "storage_document_id"),
			// 	),
			// },
			// Import State checks - import by ID (default)
			// {
			// 	ImportState:       true,
			// 	ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(cfgCreate.DeploymentResourceName),
			// 	ResourceName:      cfgCreate.DeploymentResourceName,
			// 	ImportStateVerify: true,
			// },
		},
	})
}

// testAccCheckDeploymentExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckDeploymentExists(deploymentResourceName string, deployment *api.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the deployment resource we just created from the state
		deploymentResource, exists := s.RootModule().Resources[deploymentResourceName]
		if !exists {
			return fmt.Errorf("deployment resource not found: %s", deploymentResourceName)
		}
		deploymentID, _ := uuid.Parse(deploymentResource.Primary.ID)

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("workspace resource not found: %s", testutils.WorkspaceResourceName)
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
	name                   string
	description            string
	parameterOpenapiSchema map[string]interface{}
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

		equal, diffs := helpers.ObjectsEqual(expectedValues.parameterOpenapiSchema, fetchedDeployment.ParameterOpenAPISchema)
		if !equal {
			return fmt.Errorf("Found unexpected differences in deployment parameter_openapi_schema: %s", strings.Join(diffs, "\n"))
		}

		return nil
	}
}
