provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_flow" "test" {
  name         = "my-flow"
  workspace_id = data.prefect_workspace.test.id
  tags         = ["test"]
}

resource "prefect_deployment" "test" {
  name         = "my-deployment"
  workspace_id = data.prefect_workspace.test.id
  flow_id      = prefect_flow.test.id
}

resource "prefect_deployment_schedule" "test" {
  workspace_id  = data.prefect_workspace.test.id
  deployment_id = prefect_deployment.test.id

  active          = true
  catchup         = false
  max_active_runs = 10
  timezone        = "America/New_York"

  # Option: Interval schedule
  interval    = 30
  anchor_date = "2024-01-01T00:00:00Z"

  # Option: Cron schedule
  cron   = "0 0 * * *"
  day_or = true

  # Option: RRule schedule
  rrule = "FREQ=DAILY;INTERVAL=1"
}

