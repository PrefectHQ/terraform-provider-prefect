data "prefect_global_concurrency_limit" "limit_by_id" {
  id           = "00000000-0000-0000-0000-000000000000"
  workspace_id = "00000000-0000-0000-0000-000000000000"
}

data "prefect_global_concurrency_limit" "limit_by_name" {
  name         = "my-limit"
  workspace_id = "00000000-0000-0000-0000-000000000000"
}