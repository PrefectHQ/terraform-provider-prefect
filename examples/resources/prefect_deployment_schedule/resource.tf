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

resource "prefect_deployment_schedule" "test_interval" {
  workspace_id  = data.prefect_workspace.test.id
  deployment_id = prefect_deployment.test.id

  active   = true
  catchup  = false
  timezone = "America/New_York"

  # Interval-specific fields
  interval    = 30
  anchor_date = "2024-01-01T00:00:00Z"
}

resource "prefect_deployment_schedule" "test_cron" {
  workspace_id  = data.prefect_workspace.test.id
  deployment_id = prefect_deployment.test.id

  active   = true
  catchup  = false
  timezone = "America/New_York"

  # Cron-specific fields
  cron   = "0 0 * * *"
  day_or = true
}

resource "prefect_deployment_schedule" "test_rrule" {
  workspace_id  = data.prefect_workspace.test.id
  deployment_id = prefect_deployment.test.id

  active   = true
  catchup  = false
  timezone = "America/New_York"

  # RRule-specific fields
  rrule = "FREQ=DAILY;INTERVAL=1"
}

