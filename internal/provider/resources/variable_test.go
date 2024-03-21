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

func fixtureAccVariableResource(name string, value string) string {
	return fmt.Sprintf(`
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
resource "prefect_variable" "test" {
	workspace_id = data.prefect_workspace.evergreen.id
	name = "%s"
	value = "%s"
}
	`, name, value)
}

func fixtureAccVariableResourceWithTags(name string, value string) string {
	return fmt.Sprintf(`
data "prefect_workspace" "evergreen" {
	handle = "github-ci-tests"
}
resource "prefect_variable" "test" {
	workspace_id = data.prefect_workspace.evergreen.id
	name = "%s"
	value = "%s"
	tags = ["foo", "bar"]
}
	`, name, value)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_variable(t *testing.T) {
	resourceName := "prefect_variable.test"
	const workspaceDatsourceName = "data.prefect_workspace.evergreen"

	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	randomValue := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomValue2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var variable api.Variable

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the variable resource
				Config: fixtureAccVariableResource(randomName, randomValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceDatsourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName, Value: randomValue}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue),
				),
			},
			{
				// Check updating name + value of the variable resource
				Config: fixtureAccVariableResource(randomName2, randomValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceDatsourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: randomValue2}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue2),
				),
			},
			{
				// Check adding tags
				Config: fixtureAccVariableResourceWithTags(randomName2, randomValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceDatsourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: randomValue2}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "value", randomValue2),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "bar"),
				),
			},
		},
	})
}

func testAccCheckVariableExists(variableResourceName string, workspaceDatasourceName string, variable *api.Variable) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		variableResource, exists := state.RootModule().Resources[variableResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", variableResourceName)
		}
		variableResourceID, _ := uuid.Parse(variableResource.Primary.ID)

		workspaceDatsource, exists := state.RootModule().Resources[workspaceDatasourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workspaceDatasourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceDatsource.Primary.ID)

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		variablesClient, _ := c.Variables(uuid.Nil, workspaceID)

		variableName := variableResource.Primary.Attributes["name"]

		fetchedVariable, err := variablesClient.Get(context.Background(), variableResourceID)
		if err != nil {
			return fmt.Errorf("Error fetching variable: %w", err)
		}
		if fetchedVariable == nil {
			return fmt.Errorf("Variable not found for name: %s", variableName)
		}

		*variable = *fetchedVariable

		return nil
	}
}
func testAccCheckVariableValues(fetchedVariable *api.Variable, valuesToCheck *api.Variable) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if fetchedVariable.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected variable name to be %s, got %s", valuesToCheck.Name, fetchedVariable.Name)
		}
		if fetchedVariable.Value != valuesToCheck.Value {
			return fmt.Errorf("Expected variable value to be %s, got %s", valuesToCheck.Name, fetchedVariable.Name)
		}

		return nil
	}
}
