data "prefect_work_queue" "example" {
  name           = "my-work-pool"
  work_pool_name = "my-work-queue"
  workspace_id   = prefect_workspace.example.id
}
