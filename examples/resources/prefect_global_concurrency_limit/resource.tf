provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_global_concurrency_limit" "test" {
  workspace_id          = data.prefect_workspace.test.id
  name                  = "test-global-concurrency-limit"
  limit                 = 1
  active                = true
  active_slots          = 0
  slot_decay_per_second = 1.5
}