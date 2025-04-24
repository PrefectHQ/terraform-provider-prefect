package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccTaskRunConcurrencyLimitCreate(workspace, workspaceIDArg, tag string, concurrencyLimit int64) string {
	return fmt.Sprintf(`
%s

resource "prefect_task_run_concurrency_limit" "task_run_concurrency_limit" {
	%s
	tag = "%s"
	concurrency_limit = %d
}
`, workspace, workspaceIDArg, tag, concurrencyLimit)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_task_run_concurrency_limit(t *testing.T) {
	resourceName := "prefect_task_run_concurrency_limit.task_run_concurrency_limit"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the resource
				Config: fixtureAccTaskRunConcurrencyLimitCreate(workspace.Resource, workspace.IDArg, "test1", 10),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "tag", "test1"),
					testutils.ExpectKnownValueNumber(resourceName, "concurrency_limit", 10),
				},
			},
			{
				// Check updating the resource
				Config: fixtureAccTaskRunConcurrencyLimitCreate(workspace.Resource, workspace.IDArg, "test2", 20),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "tag", "test2"),
					testutils.ExpectKnownValueNumber(resourceName, "concurrency_limit", 20),
				},
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
