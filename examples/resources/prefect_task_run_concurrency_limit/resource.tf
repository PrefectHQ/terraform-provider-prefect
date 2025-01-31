provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_task_run_concurrency_limit" "test" {
  workspace_id      = data.prefect_workspace.test.id
  concurrency_limit = 1
  tag               = "test-tag"
}

# Example of how to reference this resources is in the `examples/resources/prefect_task_run_concurrency_limit/task-run-concurrency.py` file
