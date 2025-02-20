data "prefect_work_queues" "test" {
  work_pool_name = prefect_work_pool.test.name
  workspace_id   = prefect_workspace.test.id
}
