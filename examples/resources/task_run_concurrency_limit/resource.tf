provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_task_run_concurrency_limit" "test" {
  workspace_id      = data.prefect_workspace.test.id
  concurrency_limit = 1
  tag               = "test-tag"
}
