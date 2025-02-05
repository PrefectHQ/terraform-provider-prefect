package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type automationFixtureConfig struct {
	EphemeralWorkspace             string
	EphemeralWorkspaceResourceName string
	AutomationResourceName         string
}

func fixtureAccAutomationResourceEventTrigger(cfg automationFixtureConfig) string {
	tmpl := `
{{ .EphemeralWorkspace }}

resource "prefect_automation" "{{ .AutomationResourceName }}" {
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id

	name         = "test-event-automation"
	description  = "description for test-event-automation"
	enabled      = true

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
      type = "run-deployment"
      source        = "selected"
      deployment_id = "123e4567-e89b-12d3-a456-426614174000"
      parameters = jsonencode({
        param1 = "value1"
        param2 = "value2"
      })
      job_variables = jsonencode({
        string_var = "value1"
				int_var = 2
				bool_var = true
      })
    }
  ]
}

data "prefect_automation" "{{ .AutomationResourceName }}" {
	id = prefect_automation.{{ .AutomationResourceName }}.id
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

func fixtureAccAutomationResourceMetricTrigger(cfg automationFixtureConfig) string {
	tmpl := `
{{ .EphemeralWorkspace }}

resource "prefect_automation" "{{ .AutomationResourceName }}" {
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id

	name         = "test-metric-automation"
	description  = "description for test-metric-automation"
	enabled      = true

  trigger = {
    metric = {
      match = jsonencode({
				"prefect.resource.id" = "prefect.flow-run.*"
			})
      match_related = jsonencode({
				"prefect.resource.id" = "prefect.flow-run.*"
				"prefect.resource.role" = "flow"
			})
      metric = {
        name = "duration"
        operator = ">="
        threshold = 0.5
        range = 30
        firing_for = 60
      }
    }
  }

  actions = [
    {
      type    = "change-flow-run-state"
      state   = "FAILED"
      name    = "Failed by automation"
      message = "Flow run failed"
    }
  ]
}

data "prefect_automation" "{{ .AutomationResourceName }}" {
	id = prefect_automation.{{ .AutomationResourceName }}.id
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

func fixtureAccAutomationResourceCompoundTrigger(cfg automationFixtureConfig) string {
	tmpl := `
{{ .EphemeralWorkspace }}

resource "prefect_automation" "{{ .AutomationResourceName }}" {
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id

	name         = "test-compound-automation"
	description  = "description for test-compound-automation"
	enabled      = false

  trigger = {
    compound = {
      require = "any"
      within = 302
      triggers = [
        {
          event = {
            expect = ["prefect.flow-run.Failed"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
              "prefect.resource.role" = "flow"
            })
            posture = "Reactive"
            threshold = 1
            within = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.Completed"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
              "prefect.resource.role" = "flow"
            })
            posture = "Reactive"
            threshold = 1
            within = 0
          }
        }
      ]
    }
  }

  actions = [
    {
      type = "run-deployment"
      source = "inferred"
      job_variables = jsonencode({
        var1 = "value1"
        var2 = "value2"
        var3 = "value3"
      })
    }
  ]
}

data "prefect_automation" "{{ .AutomationResourceName }}" {
	id = prefect_automation.{{ .AutomationResourceName }}.id
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

func fixtureAccAutomationResourceSequenceTrigger(cfg automationFixtureConfig) string {
	tmpl := `
{{ .EphemeralWorkspace }}

resource "prefect_automation" "{{ .AutomationResourceName }}" {
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id

	name         = "test-sequence-automation"
	description  = "description for test-sequence-automation"
	enabled      = true

  trigger = {
    sequence = {
      within = 180
      triggers = [
        {
          event = {
            expect = ["prefect.flow-run.Pending"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            posture = "Reactive"
            threshold = 1
            within = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.Running"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
							"prefect.resource.role" = "flow"
							"prefect.resource.id" = ["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"]
						})
            posture = "Reactive"
            threshold = 1
            within = 0
          }
        },
        {
          event = {
            expect = ["prefect.flow-run.Completed"]
            match = jsonencode({
              "prefect.resource.id" = "prefect.flow-run.*"
            })
            match_related = jsonencode({
							"prefect.resource.id" = ["prefect.flow-run.*"]
							"prefect.resource.role" = "flow"
						})
            posture = "Reactive"
            threshold = 1
            within = 0
          }
        }
      ]
    }
  }

  actions = [
    {
      type = "send-notification"
      block_document_id = "123e4567-e89b-12d3-a456-426614174000"
      subject = "Flow run failed"
      body = "Flow run failed at this time"
    }
  ]
}

data "prefect_automation" "{{ .AutomationResourceName }}" {
	id = prefect_automation.{{ .AutomationResourceName }}.id
	workspace_id = {{ .EphemeralWorkspaceResourceName }}.id
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_automation(t *testing.T) {
	eventTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	eventTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", eventTriggerAutomationResourceName)
	eventTriggerAutomationDataSourceNameAndPath := fmt.Sprintf("data.prefect_automation.%s", eventTriggerAutomationResourceName)

	metricTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	metricTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", metricTriggerAutomationResourceName)
	metricTriggerAutomationDataSourceNameAndPath := fmt.Sprintf("data.prefect_automation.%s", metricTriggerAutomationResourceName)

	compoundTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	compoundTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", compoundTriggerAutomationResourceName)
	compoundTriggerAutomationDataSourceNameAndPath := fmt.Sprintf("data.prefect_automation.%s", compoundTriggerAutomationResourceName)

	sequenceTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	sequenceTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", sequenceTriggerAutomationResourceName)
	sequenceTriggerAutomationDataSourceNameAndPath := fmt.Sprintf("data.prefect_automation.%s", sequenceTriggerAutomationResourceName)

	ephemeralWorkspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccAutomationResourceEventTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         eventTriggerAutomationResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(eventTriggerAutomationDataSourceNameAndPath, "id"),
					testutils.CompareValuePairs(eventTriggerAutomationDataSourceNameAndPath, "name", eventTriggerAutomationResourceNameAndPath, "name"),
					testutils.CompareValuePairs(eventTriggerAutomationDataSourceNameAndPath, "description", eventTriggerAutomationResourceNameAndPath, "description"),
					testutils.CompareValuePairs(eventTriggerAutomationDataSourceNameAndPath, "enabled", eventTriggerAutomationResourceNameAndPath, "enabled"),
					testutils.ExpectKnownValueNotNull(eventTriggerAutomationDataSourceNameAndPath, "trigger.event.posture"),
					testutils.ExpectKnownValueNull(eventTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					testutils.ExpectKnownValueNull(eventTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					testutils.ExpectKnownValueNull(eventTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					testutils.ExpectKnownValueListSize(eventTriggerAutomationDataSourceNameAndPath, "actions", 1),
					testutils.ExpectKnownValue(eventTriggerAutomationDataSourceNameAndPath, "actions.0.type", "run-deployment"),
				},
			},
			{
				Config: fixtureAccAutomationResourceMetricTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         metricTriggerAutomationResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(metricTriggerAutomationDataSourceNameAndPath, "id"),
					testutils.CompareValuePairs(metricTriggerAutomationDataSourceNameAndPath, "name", metricTriggerAutomationResourceNameAndPath, "name"),
					testutils.CompareValuePairs(metricTriggerAutomationDataSourceNameAndPath, "description", metricTriggerAutomationResourceNameAndPath, "description"),
					testutils.CompareValuePairs(metricTriggerAutomationDataSourceNameAndPath, "enabled", metricTriggerAutomationResourceNameAndPath, "enabled"),
					testutils.ExpectKnownValueNotNull(metricTriggerAutomationDataSourceNameAndPath, "trigger.metric.metric.name"),
					testutils.ExpectKnownValueNull(metricTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					testutils.ExpectKnownValueNull(metricTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					testutils.ExpectKnownValueNull(metricTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					testutils.ExpectKnownValueListSize(metricTriggerAutomationDataSourceNameAndPath, "actions", 1),
					testutils.ExpectKnownValue(metricTriggerAutomationDataSourceNameAndPath, "actions.0.type", "change-flow-run-state"),
				},
			},
			{
				Config: fixtureAccAutomationResourceCompoundTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         compoundTriggerAutomationResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(compoundTriggerAutomationDataSourceNameAndPath, "id"),
					testutils.CompareValuePairs(compoundTriggerAutomationDataSourceNameAndPath, "name", compoundTriggerAutomationResourceNameAndPath, "name"),
					testutils.CompareValuePairs(compoundTriggerAutomationDataSourceNameAndPath, "description", compoundTriggerAutomationResourceNameAndPath, "description"),
					testutils.CompareValuePairs(compoundTriggerAutomationDataSourceNameAndPath, "enabled", compoundTriggerAutomationResourceNameAndPath, "enabled"),
					testutils.ExpectKnownValueListSize(compoundTriggerAutomationDataSourceNameAndPath, "trigger.compound.triggers", 2),
					testutils.ExpectKnownValueNull(compoundTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					testutils.ExpectKnownValueNull(compoundTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					testutils.ExpectKnownValueNull(compoundTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					testutils.ExpectKnownValueListSize(compoundTriggerAutomationDataSourceNameAndPath, "actions", 1),
					testutils.ExpectKnownValue(compoundTriggerAutomationDataSourceNameAndPath, "actions.0.type", "run-deployment"),
				},
			},
			{
				Config: fixtureAccAutomationResourceSequenceTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         sequenceTriggerAutomationResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(sequenceTriggerAutomationDataSourceNameAndPath, "id"),
					testutils.CompareValuePairs(sequenceTriggerAutomationDataSourceNameAndPath, "name", sequenceTriggerAutomationResourceNameAndPath, "name"),
					testutils.CompareValuePairs(sequenceTriggerAutomationDataSourceNameAndPath, "description", sequenceTriggerAutomationResourceNameAndPath, "description"),
					testutils.CompareValuePairs(sequenceTriggerAutomationDataSourceNameAndPath, "enabled", sequenceTriggerAutomationResourceNameAndPath, "enabled"),
					testutils.ExpectKnownValueListSize(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.sequence.triggers", 3),
					testutils.ExpectKnownValueNull(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					testutils.ExpectKnownValueNull(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					testutils.ExpectKnownValueNull(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					testutils.ExpectKnownValueListSize(sequenceTriggerAutomationDataSourceNameAndPath, "actions", 1),
					testutils.ExpectKnownValue(sequenceTriggerAutomationDataSourceNameAndPath, "actions.0.type", "send-notification"),
				},
			},
		},
	})
}
