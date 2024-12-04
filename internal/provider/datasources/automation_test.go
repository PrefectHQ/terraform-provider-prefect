package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(eventTriggerAutomationDataSourceNameAndPath, "id"),
					resource.TestCheckResourceAttrPair(eventTriggerAutomationDataSourceNameAndPath, "name", eventTriggerAutomationResourceNameAndPath, "name"),
					resource.TestCheckResourceAttrPair(eventTriggerAutomationDataSourceNameAndPath, "description", eventTriggerAutomationResourceNameAndPath, "description"),
					resource.TestCheckResourceAttrPair(eventTriggerAutomationDataSourceNameAndPath, "enabled", eventTriggerAutomationResourceNameAndPath, "enabled"),
					resource.TestCheckResourceAttrSet(eventTriggerAutomationDataSourceNameAndPath, "trigger.event.posture"),
					resource.TestCheckNoResourceAttr(eventTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					resource.TestCheckNoResourceAttr(eventTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					resource.TestCheckNoResourceAttr(eventTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					resource.TestCheckResourceAttr(eventTriggerAutomationDataSourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(eventTriggerAutomationDataSourceNameAndPath, "actions.0.type", "run-deployment"),
				),
			},
			{
				Config: fixtureAccAutomationResourceMetricTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         metricTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(metricTriggerAutomationDataSourceNameAndPath, "id"),
					resource.TestCheckResourceAttrPair(metricTriggerAutomationDataSourceNameAndPath, "name", metricTriggerAutomationResourceNameAndPath, "name"),
					resource.TestCheckResourceAttrPair(metricTriggerAutomationDataSourceNameAndPath, "description", metricTriggerAutomationResourceNameAndPath, "description"),
					resource.TestCheckResourceAttrPair(metricTriggerAutomationDataSourceNameAndPath, "enabled", metricTriggerAutomationResourceNameAndPath, "enabled"),
					resource.TestCheckResourceAttrSet(metricTriggerAutomationDataSourceNameAndPath, "trigger.metric.metric.name"),
					resource.TestCheckNoResourceAttr(metricTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					resource.TestCheckNoResourceAttr(metricTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					resource.TestCheckNoResourceAttr(metricTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					resource.TestCheckResourceAttr(metricTriggerAutomationDataSourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(metricTriggerAutomationDataSourceNameAndPath, "actions.0.type", "change-flow-run-state"),
				),
			},
			{
				Config: fixtureAccAutomationResourceCompoundTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         compoundTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(compoundTriggerAutomationDataSourceNameAndPath, "id"),
					resource.TestCheckResourceAttrPair(compoundTriggerAutomationDataSourceNameAndPath, "name", compoundTriggerAutomationResourceNameAndPath, "name"),
					resource.TestCheckResourceAttrPair(compoundTriggerAutomationDataSourceNameAndPath, "description", compoundTriggerAutomationResourceNameAndPath, "description"),
					resource.TestCheckResourceAttrPair(compoundTriggerAutomationDataSourceNameAndPath, "enabled", compoundTriggerAutomationResourceNameAndPath, "enabled"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "trigger.compound.triggers.#", "2"),
					resource.TestCheckNoResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					resource.TestCheckNoResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					resource.TestCheckNoResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "trigger.sequence"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationDataSourceNameAndPath, "actions.0.type", "run-deployment"),
				),
			},
			{
				Config: fixtureAccAutomationResourceSequenceTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         sequenceTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(sequenceTriggerAutomationDataSourceNameAndPath, "id"),
					resource.TestCheckResourceAttrPair(sequenceTriggerAutomationDataSourceNameAndPath, "name", sequenceTriggerAutomationResourceNameAndPath, "name"),
					resource.TestCheckResourceAttrPair(sequenceTriggerAutomationDataSourceNameAndPath, "description", sequenceTriggerAutomationResourceNameAndPath, "description"),
					resource.TestCheckResourceAttrPair(sequenceTriggerAutomationDataSourceNameAndPath, "enabled", sequenceTriggerAutomationResourceNameAndPath, "enabled"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.sequence.triggers.#", "3"),
					resource.TestCheckNoResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.compound"),
					resource.TestCheckNoResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.event"),
					resource.TestCheckNoResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "trigger.metric"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationDataSourceNameAndPath, "actions.0.type", "send-notification"),
				),
			},
		},
	})
}
