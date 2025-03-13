resource "prefect_flow" "my_flow" {
  name = "my-flow"
}

resource "prefect_deployment" "my_deployment" {
  name    = "my-deployment"
  flow_id = prefect_flow.my_flow.id
}

resource "prefect_resource_sla" "slas" {
  resource_id = "prefect.deployment.${prefect_deployment.my_deployment.id}"
  slas = [
    {
      name     = "my-time-to-completion-sla" # can be any string
      duration = 60                          # Max flow run duration in seconds before SLA is violated
    },
    {
      name        = "my-frequency-sla"
      severity    = "critical"
      stale_after = 60 # Amount of time in seconds after a flow run is considered stale
    },
    {
      name     = "my-freshness"
      severity = "moderate"
      within   = 60 # The amount of time after a flow run is considered stale.
      resource_match = jsonencode({
        label = "my-label"
      })
      expected_event = "my-event"
    },
    {
      name     = "my-lateness"
      severity = "moderate"
      within   = 60 # The amount of time after a flow run is considered stale.
    }
  ]
}
