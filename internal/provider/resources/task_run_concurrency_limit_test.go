package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccTaskRunConcurrencyLimitCreate(workspace, tag string, concurrencyLimit int64) string {
	return fmt.Sprintf(`
%s

resource "prefect_task_run_concurrency_limit" "task_run_concurrency_limit" {
	workspace_id = prefect_workspace.test.id
	tag = "%s"
	concurrency_limit = %d
}
`, workspace, tag, concurrencyLimit)
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
				Config: fixtureAccTaskRunConcurrencyLimitCreate(workspace.Resource, "test1", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "concurrency_limit", "10"),
				),
			},
			{
				// Check updating the resource
				Config: fixtureAccTaskRunConcurrencyLimitCreate(workspace.Resource, "test2", 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tag", "test2"),
					resource.TestCheckResourceAttr(resourceName, "concurrency_limit", "20"),
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
