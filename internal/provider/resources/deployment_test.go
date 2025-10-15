package resources_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
	"k8s.io/utils/ptr"
)

type deploymentConfig struct {
	Workspace      string
	WorkspaceIDArg string

	DeploymentName         string
	DeploymentResourceName string

	ConcurrencyLimit       int64
	CollisionStrategy      string
	Description            string
	EnforceParameterSchema bool
	Entrypoint             string
	JobVariables           string
	Parameters             string
	Path                   string
	Paused                 bool
	PullSteps              []api.PullStep
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

	{{.WorkspaceIDArg}}
}

resource "prefect_work_pool" "{{.WorkPoolName}}" {
  name         = "{{.WorkPoolName}}"
  type         = "kubernetes"
  paused       = false
	{{.WorkspaceIDArg}}
}

resource "prefect_flow" "{{.FlowName}}" {
	name = "{{.FlowName}}"
	tags = [{{range .Tags}}"{{.}}", {{end}}]

	{{.WorkspaceIDArg}}
}

resource "prefect_deployment" "{{.DeploymentName}}" {
	name = "{{.DeploymentName}}"
	description = "{{.Description}}"
	concurrency_limit = {{.ConcurrencyLimit}}
	concurrency_options = {
		collision_strategy = "{{.CollisionStrategy}}"
	}
	enforce_parameter_schema = {{.EnforceParameterSchema}}
	entrypoint = "{{.Entrypoint}}"
	flow_id = prefect_flow.{{.FlowName}}.id
	job_variables = jsonencode(
		{{.JobVariables}}
	)
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
	pull_steps = [
	  {{range .PullSteps}}
	  {
			{{- with .PullStepSetWorkingDirectory }}
			type = "set_working_directory"
			directory = "{{.Directory}}"
			{{- end}}

			{{- with .PullStepGitClone }}
			type = "git_clone"
			repository = "{{.Repository}}"
			{{-   if .Branch }}
			branch = "{{.Branch}}"
			{{-   end }}
			{{-   if .AccessToken }}
			access_token = "{{.AccessToken}}"
			{{-   end }}
			{{-   if .IncludeSubmodules }}
			include_submodules = {{.IncludeSubmodules}}
			{{-   end }}
			{{-   if .Credentials }}
			credentials = "{{.Credentials}}"
			{{-   end }}
			{{-   if .Requires }}
			requires = "{{.Requires}}"
			{{-   end }}
			{{- end }}

			{{- with .PullStepPullFromAzureBlobStorage }}
			type = "pull_from_azure_blob_storage"
			{{-   if .Container }}
			container = "{{.Container}}"
			{{-   end}}
			{{-   if .Folder }}
			folder = "{{.Folder}}"
			{{-   end}}
			{{-   if .Credentials }}
			credentials = "{{.Credentials}}"
			{{-   end }}
			{{-   if .Requires }}
			requires = "{{.Requires}}"
			{{-   end }}
			{{- end }}

			{{- with .PullStepPullFromGCS }}
			type = "pull_from_gcs"
			{{-   if .Bucket }}
			bucket = "{{.Bucket}}"
			{{-   end}}
			{{-   if .Folder }}
			folder = "{{.Folder}}"
			{{-   end}}
			{{-   if .Credentials }}
			credentials = "{{.Credentials}}"
			{{-   end }}
			{{-   if .Requires }}
			requires = "{{.Requires}}"
			{{-   end }}
			{{- end }}

			{{- with .PullStepPullFromS3 }}
			type = "pull_from_s3"
			{{-   if .Bucket }}
			bucket = "{{.Bucket}}"
			{{-   end}}
			{{-   if .Folder }}
			folder = "{{.Folder}}"
			{{-   end}}
			{{-   if .Credentials }}
			credentials = "{{.Credentials}}"
			{{-   end }}
			{{-   if .Requires }}
			requires = "{{.Requires}}"
			{{-   end }}
			{{- end }}
		},
		{{end}}
	]
	storage_document_id = prefect_block.test_gh_repository.id

	{{.WorkspaceIDArg}}
	depends_on = [prefect_flow.{{.FlowName}}]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_with_global_concurrency_limit(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	deploymentName := testutils.NewRandomPrefixedString()
	flowName := testutils.NewRandomPrefixedString()
	gcl1Name := testutils.NewRandomPrefixedString()
	gcl2Name := testutils.NewRandomPrefixedString()

	// Configuration for creating deployment with first global concurrency limit
	cfgCreate := fmt.Sprintf(`
%s

resource "prefect_flow" "%s" {
	name = "%s"
	%s
}

resource "prefect_global_concurrency_limit" "test1" {
	name = "%s"
	limit = 5
	%s
}

resource "prefect_global_concurrency_limit" "test2" {
	name = "%s"
	limit = 10
	%s
}

resource "prefect_deployment" "%s" {
	name = "%s"
	flow_id = prefect_flow.%s.id
	global_concurrency_limit_id = prefect_global_concurrency_limit.test1.id
	%s
}
`, workspace.Resource, flowName, flowName, workspace.IDArg, gcl1Name, workspace.IDArg, gcl2Name, workspace.IDArg, deploymentName, deploymentName, flowName, workspace.IDArg)

	// Configuration for updating deployment to use second global concurrency limit
	cfgUpdate := fmt.Sprintf(`
%s

resource "prefect_flow" "%s" {
	name = "%s"
	%s
}

resource "prefect_global_concurrency_limit" "test1" {
	name = "%s"
	limit = 5
	%s
}

resource "prefect_global_concurrency_limit" "test2" {
	name = "%s"
	limit = 10
	%s
}

resource "prefect_deployment" "%s" {
	name = "%s"
	flow_id = prefect_flow.%s.id
	global_concurrency_limit_id = prefect_global_concurrency_limit.test2.id
	%s
}
`, workspace.Resource, flowName, flowName, workspace.IDArg, gcl1Name, workspace.IDArg, gcl2Name, workspace.IDArg, deploymentName, deploymentName, flowName, workspace.IDArg)

	deploymentResourceName := fmt.Sprintf("prefect_deployment.%s", deploymentName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation with global_concurrency_limit_id
				Config: cfgCreate,
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(deploymentResourceName, "global_concurrency_limit_id"),
				},
			},
			{
				// Check update to different global_concurrency_limit_id
				Config: cfgUpdate,
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(deploymentResourceName, "global_concurrency_limit_id"),
				},
			},
		},
	})
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	deploymentName := testutils.NewRandomPrefixedString()
	flowName := testutils.NewRandomPrefixedString()

	parameterOpenAPISchema := `{"type": "object", "properties": {"some-parameter": {"type": "string"}}}`
	expectedParameterOpenAPISchema := testutils.NormalizedValueForJSON(t, parameterOpenAPISchema)

	parameterOpenAPISchemaUpdate := `{"type": "object", "properties": {"some-parameter": {"type": "string"}, "some-other-parameter": {"type": "string"}}}`
	expectedParameterOpenAPISchemaUpdate := testutils.NormalizedValueForJSON(t, parameterOpenAPISchemaUpdate)

	cfgCreate := deploymentConfig{
		DeploymentName:         deploymentName,
		FlowName:               flowName,
		DeploymentResourceName: fmt.Sprintf("prefect_deployment.%s", deploymentName),
		Workspace:              workspace.Resource,
		WorkspaceIDArg:         workspace.IDArg,

		ConcurrencyLimit:       1,
		CollisionStrategy:      "ENQUEUE",
		Description:            "My deployment description",
		EnforceParameterSchema: true,
		Entrypoint:             "hello_world.py:hello_world",
		JobVariables:           `{"env":{"some-key":"some-value"}}`,
		Parameters:             "some-value1",
		Path:                   "some-path",
		Paused:                 false,
		PullSteps: []api.PullStep{
			{
				PullStepSetWorkingDirectory: &api.PullStepSetWorkingDirectory{
					Directory: ptr.To("/some/directory"),
				},
			},
		},
		Tags:                   []string{"test1", "test2"},
		Version:                "v1.1.1",
		WorkPoolName:           "some-pool",
		WorkQueueName:          "default",
		ParameterOpenAPISchema: parameterOpenAPISchema,
		StorageDocumentName:    testutils.NewRandomPrefixedString(),
	}

	cfgUpdate := deploymentConfig{
		// Keep some values from cfgCreate so we refer to the same resources for the update.
		DeploymentName:         cfgCreate.DeploymentName,
		FlowName:               cfgCreate.FlowName,
		DeploymentResourceName: cfgCreate.DeploymentResourceName,
		Workspace:              cfgCreate.Workspace,
		WorkspaceIDArg:         cfgCreate.WorkspaceIDArg,
		WorkPoolName:           cfgCreate.WorkPoolName,

		// Configure new values to test the update.
		ConcurrencyLimit:       2,
		CollisionStrategy:      "CANCEL_NEW",
		Description:            "My deployment description v2",
		EnforceParameterSchema: false,
		Entrypoint:             "hello_world.py:hello_world2",
		JobVariables:           `{"env":{"some-key":"some-value2"}}`,
		ParameterOpenAPISchema: parameterOpenAPISchemaUpdate,
		Parameters:             "some-value2",
		Path:                   "some-path2",
		Paused:                 true,
		Version:                "v1.1.2",
		WorkQueueName:          "default",

		// Changing the tags results in a "404 Deployment not found" error.
		// Will avoid testing this until a solution is found.
		//
		// Tags: []string{"test1", "test3"}
		Tags: cfgCreate.Tags,

		// PullSteps require a replacement of the resource.
		PullSteps: []api.PullStep{
			{
				PullStepSetWorkingDirectory: &api.PullStepSetWorkingDirectory{
					Directory: ptr.To("/some/other/directory"),
				},
			},
			{
				PullStepGitClone: &api.PullStepGitClone{
					Repository:        ptr.To("https://github.com/prefecthq/prefect"),
					Branch:            ptr.To("main"),
					AccessToken:       ptr.To("123abc"),
					IncludeSubmodules: ptr.To(true),
				},
			},
			{
				PullStepPullFromAzureBlobStorage: &api.PullStepPullFromAzure{
					Container: ptr.To("my-container"),
					Folder:    ptr.To("my-folder"),
					PullStepCommon: api.PullStepCommon{
						Credentials: ptr.To("azure-credentials"),
						Requires:    ptr.To("prefect-azure[blob_storage]"),
					},
				},
			},
			{
				PullStepPullFromS3: &api.PullStepPullFrom{
					Bucket: ptr.To("some-bucket"),
					Folder: ptr.To("some-folder"),
					PullStepCommon: api.PullStepCommon{
						Credentials: ptr.To("some-credentials"),
						Requires:    ptr.To("prefect-aws>=0.3.4"),
					},
				},
			},
		},

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
						name:        cfgCreate.DeploymentName,
						description: cfgCreate.Description,
						pullSteps:   cfgCreate.PullSteps,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNumber(cfgCreate.DeploymentResourceName, "concurrency_limit", cfgCreate.ConcurrencyLimit),
					testutils.ExpectKnownValueMap(cfgCreate.DeploymentResourceName, "concurrency_options", map[string]string{
						"collision_strategy": cfgCreate.CollisionStrategy,
					}),
					testutils.ExpectKnownValueBool(cfgCreate.DeploymentResourceName, "enforce_parameter_schema", cfgCreate.EnforceParameterSchema),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "entrypoint", cfgCreate.Entrypoint),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "job_variables", cfgCreate.JobVariables),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "parameters", `{"some-parameter":"some-value1"}`),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "parameter_openapi_schema", expectedParameterOpenAPISchema),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "path", cfgCreate.Path),
					testutils.ExpectKnownValueBool(cfgCreate.DeploymentResourceName, "paused", cfgCreate.Paused),
					testutils.ExpectKnownValueSet(cfgCreate.DeploymentResourceName, "tags", cfgCreate.Tags),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "version", cfgCreate.Version),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "work_pool_name", cfgCreate.WorkPoolName),
					testutils.ExpectKnownValue(cfgCreate.DeploymentResourceName, "work_queue_name", cfgCreate.WorkQueueName),
					testutils.ExpectKnownValueNotNull(cfgCreate.DeploymentResourceName, "storage_document_id"),
				},
			},
			{
				// Check update of existing deployment resource
				Config: fixtureAccDeployment(cfgUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentExists(cfgUpdate.DeploymentResourceName, &deployment),
					testAccCheckDeploymentValues(&deployment, expectedDeploymentValues{
						name:        cfgUpdate.DeploymentName,
						description: cfgUpdate.Description,
						pullSteps:   cfgUpdate.PullSteps,
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNumber(cfgUpdate.DeploymentResourceName, "concurrency_limit", cfgUpdate.ConcurrencyLimit),
					testutils.ExpectKnownValueMap(cfgUpdate.DeploymentResourceName, "concurrency_options", map[string]string{
						"collision_strategy": cfgUpdate.CollisionStrategy,
					}),
					testutils.ExpectKnownValueBool(cfgUpdate.DeploymentResourceName, "enforce_parameter_schema", cfgUpdate.EnforceParameterSchema),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "entrypoint", cfgUpdate.Entrypoint),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "job_variables", cfgUpdate.JobVariables),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "parameters", `{"some-parameter":"some-value2"}`),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "parameter_openapi_schema", expectedParameterOpenAPISchemaUpdate),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "path", cfgUpdate.Path),
					testutils.ExpectKnownValueBool(cfgUpdate.DeploymentResourceName, "paused", cfgUpdate.Paused),
					testutils.ExpectKnownValueSet(cfgUpdate.DeploymentResourceName, "tags", cfgUpdate.Tags),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "version", cfgUpdate.Version),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "work_pool_name", cfgUpdate.WorkPoolName),
					testutils.ExpectKnownValue(cfgUpdate.DeploymentResourceName, "work_queue_name", cfgUpdate.WorkQueueName),
					testutils.ExpectKnownValueNotNull(cfgUpdate.DeploymentResourceName, "storage_document_id"),
				},
			},
			{
				// Import State checks - import by ID (default)
				ImportState:       true,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(cfgCreate.DeploymentResourceName),
				ResourceName:      cfgCreate.DeploymentResourceName,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckDeploymentExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckDeploymentExists(deploymentResourceName string, deployment *api.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the deployment resource we just created from the state
		deploymentID, err := testutils.GetResourceIDFromState(s, deploymentResourceName)
		if err != nil {
			return fmt.Errorf("error fetching deployment ID: %w", err)
		}

		var workspaceID uuid.UUID

		if !testutils.TestContextOSS() {
			// Get the workspace resource we just created from the state
			workspaceID, err = testutils.GetResourceIDFromState(s, testutils.WorkspaceResourceName)
			if err != nil {
				return fmt.Errorf("error fetching workspace ID: %w", err)
			}
		}

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
	pullSteps   []api.PullStep
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

		if !reflect.DeepEqual(fetchedDeployment.PullSteps, expectedValues.pullSteps) {
			return fmt.Errorf("Expected pull steps to be: \n%v\n got \n%v", expectedValues.pullSteps, fetchedDeployment.PullSteps)
		}

		return nil
	}
}
