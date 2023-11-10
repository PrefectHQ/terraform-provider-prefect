package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccVariableByName(name string) string {
	return fmt.Sprintf(`
	data "prefect_workspace" "evergreen" {
		handle = "evergreen-workspace"
	}
	data "prefect_variable" "test" {
		name = "%s"
		workspace_id = data.prefect_workspace.evergreen.id
	}
	`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_variable(t *testing.T) {
	datasourceName := "data.prefect_variable.test"
	variableName := "my_variable"
	variableValue := "variable value goes here"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccVariableByName(variableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "name", variableName),
					resource.TestCheckResourceAttr(datasourceName, "value", variableValue),
				),
			},
		},
	})
}
