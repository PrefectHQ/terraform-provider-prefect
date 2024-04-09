// test
package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccTeam(name string) string {
	return fmt.Sprintf(`
data "prefect_team" "default" {
	name = "%s"
}
	`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_team(t *testing.T) {
	dataSourceName := "data.prefect_team.default"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccTeam("my-team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", "my-team"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
				),
			},
		},
	})
}
