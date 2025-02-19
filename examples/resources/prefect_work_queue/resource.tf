resource "prefect_workspace" "example" {
  handle = "my-workspace"
  name   = "my-workspace"
}

resource "prefect_work_pool" "example" {
  name         = "my-work-pool"
  type         = "kubernetes"
  paused       = false
  workspace_id = prefect_workspace.example.id
}

resource "prefect_work_queue" "example" {
  name              = "my-work-queue"
  description       = "My work queue"
  concurrency_limit = 2
  workspace_id      = prefect_workspace.example.id
}
