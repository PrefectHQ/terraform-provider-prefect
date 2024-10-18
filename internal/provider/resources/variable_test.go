package resources_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccVariableResource(workspace, workspaceName, name string, value interface{}) string {
	return fmt.Sprintf(`
%s
resource "prefect_variable" "test" {
	name = "%s"
	value = %v
	workspace_id = prefect_workspace.%s.id
	depends_on = [prefect_workspace.%s]
}
	`, workspace, name, value, workspaceName, workspaceName)
}

func fixtureAccVariableResourceWithTags(workspace, workspaceName, name string, value interface{}) string {
	return fmt.Sprintf(`
%s
resource "prefect_variable" "test" {
	name = "%s"
	value = %v
	tags = ["foo", "bar"]
	workspace_id = prefect_workspace.%s.id
	depends_on = [prefect_workspace.%s]
}
	`, workspace, name, value, workspaceName, workspaceName)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_variable(t *testing.T) {
	randomName := testutils.NewRandomPrefixedString()
	randomName2 := testutils.NewRandomPrefixedString()

	resourceName := "prefect_variable.test"

	valueString := "hello-world"
	valueStringForResource := fmt.Sprintf("%q", valueString)
	valueNumber := float64(123)
	valueBool := true
	valueObject := `{"foo" = "bar"}`
	valueObjectForResource := fmt.Sprintf("%q", valueObject)

	workspace, workspaceName := testutils.NewEphemeralWorkspace()
	workspaceResourceName := "prefect_workspace." + workspaceName

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var variable api.Variable

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the variable resource
				Config: fixtureAccVariableResource(workspace, workspaceName, randomName, valueStringForResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName, Value: valueString}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
				),
			},
			{
				// Check updating name of the variable resource
				Config: fixtureAccVariableResource(workspace, workspaceName, randomName2, valueStringForResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueString}),
				),
			},
			{
				// Check updating value of the variable resource to a number
				Config: fixtureAccVariableResource(workspace, workspaceName, randomName2, valueNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueNumber}),
				),
			},
			{
				// Check updating value of the variable resource to a boolean
				Config: fixtureAccVariableResource(workspace, workspaceName, randomName2, valueBool),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueBool}),
				),
			},
			{
				// Check updating value of the variable resource to a object
				Config: fixtureAccVariableResource(workspace, workspaceName, randomName2, valueObjectForResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueObject}),
				),
			},
			{
				// Check adding tags
				Config: fixtureAccVariableResourceWithTags(workspace, workspaceName, randomName2, valueBool),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, workspaceResourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueBool}),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "bar"),
				),
			},
		},
	})
}

func testAccCheckVariableExists(variableResourceName string, workspaceResourceName string, variable *api.Variable) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		variableResource, exists := state.RootModule().Resources[variableResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", variableResourceName)
		}
		variableResourceID, _ := uuid.Parse(variableResource.Primary.ID)

		workspaceResource, exists := state.RootModule().Resources[workspaceResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

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
	return func(_ *terraform.State) error {
		if fetchedVariable.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected variable name to be %s, got %s", valuesToCheck.Name, fetchedVariable.Name)
		}

		if !reflect.DeepEqual(fetchedVariable.Value, valuesToCheck.Value) {
			return fmt.Errorf("Expected variable value to be %s, got %s", valuesToCheck.Value, fetchedVariable.Value)
		}

		return nil
	}
}
