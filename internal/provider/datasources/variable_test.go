package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccVariableByName(workspace, name string) string {
	return fmt.Sprintf(`
	%s

	resource "prefect_variable" "test" {
		name = "%s"
		value = "variable value goes here"
		workspace_id = prefect_workspace.test.id
		depends_on = [prefect_workspace.test]
	}

	data "prefect_variable" "test" {
		name = "%s"
		workspace_id = prefect_workspace.test.id
	}
	`, workspace, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_variable(t *testing.T) {
	datasourceName := "data.prefect_variable.test"
	variableName := "my_variable"
	variableValue := "variable value goes here"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccVariableByName(workspace.Resource, variableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "name", variableName),
					resource.TestCheckResourceAttr(datasourceName, "value", variableValue),
				),
			},
		},
	})
}
