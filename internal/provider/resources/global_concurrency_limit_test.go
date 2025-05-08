package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccGlobalConcurrencyLimitCreate(workspace, workspaceIDArg, name string, limit int64, active bool, activeSlots int64, slotDecayPerSecond float64) string {
	return fmt.Sprintf(`
%s
resource "prefect_global_concurrency_limit" "global_concurrency_limit" {
	%s
	name = "%s"
	limit = %d
	active = %t
	active_slots = %d
	slot_decay_per_second = %f
}
`, workspace, workspaceIDArg, name, limit, active, activeSlots, slotDecayPerSecond)
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
				Config: fixtureAccGlobalConcurrencyLimitCreate(workspace.Resource, workspace.IDArg, "test1", 10, true, 0, 1.5),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", "test1"),
					testutils.ExpectKnownValueNumber(resourceName, "limit", 10),
					testutils.ExpectKnownValueBool(resourceName, "active", true),
					testutils.ExpectKnownValueNumber(resourceName, "active_slots", 0),
					testutils.ExpectKnownValueFloat(resourceName, "slot_decay_per_second", 1.5),
				},
			},
			{
				// Check updating the resource
				Config: fixtureAccGlobalConcurrencyLimitCreate(workspace.Resource, workspace.IDArg, "test2", 20, false, 1, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", "test2"),
					testutils.ExpectKnownValueNumber(resourceName, "limit", 20),
					testutils.ExpectKnownValueBool(resourceName, "active", false),
					testutils.ExpectKnownValueNumber(resourceName, "active_slots", 1),
					testutils.ExpectKnownValueFloat(resourceName, "slot_decay_per_second", 2),
				},
			},
			{
				// Import State checks - import by ID (default)
				ImportState:       true,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(resourceName),
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}
