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
)

func fixtureAccVariableResource(workspace, workspaceIDArg, name string, value any) string {
	return fmt.Sprintf(`
%s

resource "prefect_variable" "test" {
	name = "%s"
	value = %v
	%s
}
	`, workspace, name, value, workspaceIDArg)
}

func fixtureAccVariableResourceWithTags(workspace, workspaceIDARg, name string, value any) string {
	return fmt.Sprintf(`
%s

resource "prefect_variable" "test" {
	name = "%s"
	value = %v
	tags = ["foo", "bar"]
	%s
}
	`, workspace, name, value, workspaceIDARg)
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
	valueTuple := `["foo", "bar"]`
	valueTupleExpected := []any{`"foo"`, `"bar"`}
	valueObject := `{"foo" = "bar"}`
	valueObjectExpected := map[string]any{"foo": "bar"}

	workspace := testutils.NewEphemeralWorkspace()

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var variable api.Variable

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the variable resource
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName, valueStringForResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName, Value: valueString}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
				},
			},
			{
				// Check updating name of the variable resource
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName2, valueStringForResource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueString}),
				),
			},
			{
				// Check updating value of the variable resource to a number
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName2, valueNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueNumber}),
				),
			},
			{
				// Check updating value of the variable resource to a boolean
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName2, valueBool),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueBool}),
				),
			},
			{
				// Check updating value of the variable resource to a object
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName2, valueObject),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueObjectExpected}),
				),
			},
			{
				// Check updating value of the variable resource to a tuple
				Config: fixtureAccVariableResource(workspace.Resource, workspace.IDArg, randomName2, valueTuple),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueTupleExpected}),
				),
			},
			{
				// Check adding tags
				Config: fixtureAccVariableResourceWithTags(workspace.Resource, workspace.IDArg, randomName2, valueBool),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVariableExists(resourceName, &variable),
					testAccCheckVariableValues(&variable, &api.Variable{Name: randomName2, Value: valueBool}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName2),
					testutils.ExpectKnownValueSet(resourceName, "tags", []string{"foo", "bar"}),
				},
			},
			{
				ImportState:             true,
				ResourceName:            resourceName,
				ImportStateIdFunc:       testutils.GetResourceWorkspaceImportStateID(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func testAccCheckVariableExists(variableResourceName string, variable *api.Variable) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		variableResourceID, err := testutils.GetResourceIDFromState(state, variableResourceName)
		if err != nil {
			return fmt.Errorf("error fetching variable ID: %w", err)
		}

		var workspaceID uuid.UUID

		if !testutils.TestContextOSS() {
			// Get the workspace resource we just created from the state
			workspaceID, err = testutils.GetResourceWorkspaceIDFromState(state)
			if err != nil {
				return fmt.Errorf("error fetching workspace ID: %w", err)
			}
		}

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		variablesClient, _ := c.Variables(uuid.Nil, workspaceID)

		fetchedVariable, err := variablesClient.Get(context.Background(), variableResourceID)
		if err != nil {
			return fmt.Errorf("Error fetching variable: %w", err)
		}
		if fetchedVariable == nil {
			return fmt.Errorf("Variable not found for id: %s", variableResourceID)
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
