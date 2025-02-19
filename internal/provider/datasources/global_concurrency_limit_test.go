package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccGlobalConcurrencyLimitDataSource(workspace, name string) string {
	return fmt.Sprintf(`
%s

resource "prefect_global_concurrency_limit" "global_concurrency_limit" {
	workspace_id = prefect_workspace.test.id
	name = "%s"
	limit = 10
	active = true
	active_slots = 10
	slot_decay_per_second = 1
}
data "prefect_global_concurrency_limit" "limit_by_name" {
	name = prefect_global_concurrency_limit.global_concurrency_limit.name
	workspace_id = prefect_global_concurrency_limit.global_concurrency_limit.workspace_id
}
	`, workspace, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_global_concurrency_limit(t *testing.T) {
	dataSourceNameByName := "data.prefect_global_concurrency_limit.limit_by_name"
	workspace := testutils.NewEphemeralWorkspace()
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccGlobalConcurrencyLimitDataSource(workspace.Resource, randomName),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check the prefect_global_concurrency_limit datasource that queries by name
					testutils.ExpectKnownValue(dataSourceNameByName, "name", randomName),
					testutils.ExpectKnownValueNotNull(dataSourceNameByName, "created"),
					testutils.ExpectKnownValueNotNull(dataSourceNameByName, "updated"),
					testutils.ExpectKnownValueNumber(dataSourceNameByName, "limit", 10),
					testutils.ExpectKnownValueBool(dataSourceNameByName, "active", true),
					testutils.ExpectKnownValueNumber(dataSourceNameByName, "active_slots", 10),
					testutils.ExpectKnownValueFloat(dataSourceNameByName, "slot_decay_per_second", 1),
				},
			},
		},
	})
}
