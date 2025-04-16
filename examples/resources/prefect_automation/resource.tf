# example:
# An automation with an event trigger
resource "prefect_automation" "event_trigger" {
  name        = "tfp-test-event-trigger"
  description = "ayu carumba"
  enabled     = true

  trigger = {
    event = {
      posture = "Reactive"
      match = jsonencode({
        "prefect.resource.id" : "prefect.flow-run.*"
      })
      match_related = jsonencode({
        "prefect.resource.id" : ["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"]
        "prefect.resource.role" : "flow"
      })
      after     = ["prefect.flow-run.completed"]
      expect    = ["prefect.flow-run.failed"]
      for_each  = ["prefect.resource.id"]
      threshold = 1
      within    = 60
    }
  }
  actions = [
    {
      type   = "run-deployment"
      source = "inferred"
      parameters = jsonencode({
        "param1" = "value1"
        "param2" = "value2"
      })
      job_variables = jsonencode({
        "var1" = "value1"
        "var2" = "value2"
        "var3" = 3
        "var4" = true
        "var5" = {
          "key1" = "value1"
        }
      })
    },
    {
      type = "declare-incident"
    }
  ]
}

# example:
# An automation with a metric trigger
resource "prefect_automation" "metric_trigger" {
  name        = "tfp-test-metric-trigger"
  description = "boom shakkala"
  enabled     = true

  trigger = {
    metric = {
      posture       = "Metric"
      match         = jsonencode({})
      match_related = jsonencode({})
      metric = {
        name       = "duration"
        operator   = ">="
        threshold  = 10
        range      = 300
        firing_for = 300
      }
    }
  }
  actions = [
    {
      type    = "change-flow-run-state"
      state   = "FAILED"
      name    = "Failed by automation"
      message = "Flow run failed due to {{ event.reason }}"
    }
  ]
  actions_on_trigger = []
  actions_on_resolve = []
}

# example:
# An automation with a compound trigger
resource "prefect_automation" "compound_trigger" {
  name        = "tfp-test-compound-trigger"
  description = "compound trigger dos!"
  enabled     = true

  trigger = {
    compound = {
      require = "any"
      within  = 300
      triggers = [
        {
          event = {
            expect = ["prefect.flow-run.Failed"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
              "prefect.resource.id"   = "prefect.flow-run.*"
              "prefect.resource.role" = "flow"
            })
            for_each  = []
            after     = []
            posture   = "Reactive"
            threshold = 1
            within    = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.NonExistent"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
              "prefect.resource.id"   = "prefect.flow-run.*"
              "prefect.resource.role" = "flow"
            })
            # for_each = []
            # after = []
            posture   = "Reactive"
            threshold = 1
            within    = 0
          }
        }
      ]
    }
  }

  actions = [
    {
      type   = "run-deployment"
      source = "inferred"
      job_variables = jsonencode({
        "var1" = "value1"
        "var2" = "value2"
        "var3" = 3
        "var4" = true
        "var5" = {
          "key1" = "value1"
        }
      })
    }
  ]
}

# example:
# An automation with a sequence trigger
resource "prefect_automation" "sequence_trigger" {
  name        = "tfp-test-sequence-trigger"
  description = "sequence trigger tres!"
  enabled     = true

  trigger = {
    sequence = {
      within = 300
      triggers = [
        {
          event = {
            expect = ["prefect.flow-run.Pending"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({})
            for_each      = []
            posture       = "Reactive"
            threshold     = 1
            within        = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.Running"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({})
            for_each      = []
            posture       = "Reactive"
            threshold     = 1
            within        = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.Completed"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({})
            for_each      = []
            posture       = "Reactive"
            threshold     = 1
            within        = 0
          }
        }
      ]
    }
  }

  actions = [
    {
      type              = "send-notification"
      block_document_id = "123e4567-e89b-12d3-a456-426614174000"
      subject           = "Flow Run Failed: {{ event.resource['prefect.resource.name'] }}"
      body              = "Flow run {{ event.resource['prefect.resource.id'] }} failed at {{ event.occurred }}"
    }
  ]
  actions_on_trigger = [
    {
      type    = "change-flow-run-state"
      state   = "FAILED"
      name    = "Failed by automation"
      message = "Flow run failed due to {{ event.resource['prefect.resource.name'] }}"
    }
  ]
  actions_on_resolve = [
    {
      type              = "call-webhook"
      block_document_id = "123e4567-e89b-12d3-a456-426614174000"
      payload           = "{\"flow_run_id\": \"{{ event.resource['prefect.resource.id'] }}\", \"status\": \"{{ event.event }}\"}"
    }
  ]
}
