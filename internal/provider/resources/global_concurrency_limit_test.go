package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccGlobalConcurrencyLimitCreate(workspace, name string, limit int64, active bool, activeSlots int64, deniedSlots int64, slotDecayPerSecond int64) string {
	return fmt.Sprintf(`
%s

resource "prefect_global_concurrency_limit" "global_concurrency_limit" {
	workspace_id = prefect_workspace.test.id
	name = "%s"
	limit = %d
	active = %t
	active_slots = %d
	denied_slots = %d
	slot_decay_per_second = %d
}
`, workspace, name, limit, active, activeSlots, deniedSlots, slotDecayPerSecond)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_global_concurrency_limit(t *testing.T) {
	resourceName := "prefect_global_concurrency_limit.global_concurrency_limit"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the resource
				Config: fixtureAccGlobalConcurrencyLimitCreate(workspace.Resource, "test1", 10, true, 0, 0, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "limit", "10"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "active_slots", "0"),
					resource.TestCheckResourceAttr(resourceName, "denied_slots", "0"),
					resource.TestCheckResourceAttr(resourceName, "slot_decay_per_second", "0"),
				),
			},
			// Check updating the resource
			{
				Config: fixtureAccGlobalConcurrencyLimitCreate(workspace.Resource, "test2", 20, false, 1, 1, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "limit", "20"),
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "active_slots", "1"),
					resource.TestCheckResourceAttr(resourceName, "denied_slots", "1"),
					resource.TestCheckResourceAttr(resourceName, "slot_decay_per_second", "1"),
				),
			},
			// Import State checks - import by ID (default)
			{
				ImportState:       true,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(resourceName),
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}
