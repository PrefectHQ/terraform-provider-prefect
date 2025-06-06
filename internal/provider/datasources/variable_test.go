package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccVariableByName(workspace, workspaceIDArg, name string) string {
	return fmt.Sprintf(`
%s

resource "prefect_variable" "test" {
	name = "%s"
	value = "variable value goes here"
	%s
}

data "prefect_variable" "test" {
	name = "%s"
	%s
	depends_on = [prefect_variable.test]
}
	`, workspace, name, workspaceIDArg, name, workspaceIDArg)
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
				Config: fixtureAccVariableByName(workspace.Resource, workspace.IDArg, variableName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceName, "id"),
					testutils.ExpectKnownValue(datasourceName, "name", variableName),
					testutils.ExpectKnownValue(datasourceName, "value", variableValue),
				},
			},
		},
	})
}
