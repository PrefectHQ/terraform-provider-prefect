package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
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
      after     = [
				"prefect.flow-run.Completed",
				"prefect.flow-run.Succeeded",
			]
      expect    = [
				"prefect.flow-run.Failed",
				"prefect.flow-run.Cancelled",
				"prefect.flow-run.Crashed",
			]
      for_each  = [
				"prefect.resource.id",
				"prefect.resource.role",
			]
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
            after = [
              "prefect.flow-run.Completed",
              "prefect.flow-run.Succeeded",
            ]
            expect = [
              "prefect.flow-run.Failed",
              "prefect.flow-run.Cancelled",
              "prefect.flow-run.Crashed",
            ]
            for_each = [
              "prefect.resource.id",
              "prefect.resource.role",
            ]
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
`

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_automation(t *testing.T) {
	eventTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	eventTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", eventTriggerAutomationResourceName)

	metricTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	metricTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", metricTriggerAutomationResourceName)

	compoundTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	compoundTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", compoundTriggerAutomationResourceName)

	sequenceTriggerAutomationResourceName := testutils.NewRandomPrefixedString()
	sequenceTriggerAutomationResourceNameAndPath := fmt.Sprintf("prefect_automation.%s", sequenceTriggerAutomationResourceName)
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
					testAccCheckAutomationResourceExists(eventTriggerAutomationResourceNameAndPath, &api.Automation{}),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "name", "test-event-automation"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "description", "description for test-event-automation"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "enabled", "true"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.after.#", "2"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.after.0", "prefect.flow-run.Completed"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.after.1", "prefect.flow-run.Succeeded"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.expect.#", "3"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.expect.0", "prefect.flow-run.Cancelled"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.expect.1", "prefect.flow-run.Crashed"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.expect.2", "prefect.flow-run.Failed"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.for_each.#", "2"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.for_each.0", "prefect.resource.id"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.for_each.1", "prefect.resource.role"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.threshold", "1"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.within", "60"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.0.type", "run-deployment"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.0.source", "selected"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.0.deployment_id", "123e4567-e89b-12d3-a456-426614174000"),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "trigger.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"],"prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.0.parameters", testutils.NormalizedValueForJSON(t, `{"param1":"value1","param2":"value2"}`)),
					resource.TestCheckResourceAttr(eventTriggerAutomationResourceNameAndPath, "actions.0.job_variables", testutils.NormalizedValueForJSON(t, `{"string_var":"value1","int_var":2,"bool_var":true}`)),
				),
			},
			// Import State checks - import by automation_id
			{
				ImportState:       true,
				ResourceName:      eventTriggerAutomationResourceNameAndPath,
				ImportStateIdFunc: getAutomationImportStateID(eventTriggerAutomationResourceNameAndPath),
				ImportStateVerify: true,
			},
			{
				Config: fixtureAccAutomationResourceMetricTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         metricTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutomationResourceExists(metricTriggerAutomationResourceNameAndPath, &api.Automation{}),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "name", "test-metric-automation"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "description", "description for test-metric-automation"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "enabled", "true"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.name", "duration"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.operator", ">="),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.threshold", "0.5"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.range", "30"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.firing_for", "60"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "actions.0.type", "change-flow-run-state"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "actions.0.state", "FAILED"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "actions.0.name", "Failed by automation"),
					resource.TestCheckResourceAttr(metricTriggerAutomationResourceNameAndPath, "actions.0.message", "Flow run failed"),
				),
			},
			// Import State checks - import by automation_id
			{
				ImportState:       true,
				ResourceName:      metricTriggerAutomationResourceNameAndPath,
				ImportStateIdFunc: getAutomationImportStateID(metricTriggerAutomationResourceNameAndPath),
				ImportStateVerify: true,
			},
			{
				Config: fixtureAccAutomationResourceCompoundTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         compoundTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutomationResourceExists(compoundTriggerAutomationResourceNameAndPath, &api.Automation{}),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "name", "test-compound-automation"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "description", "description for test-compound-automation"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "enabled", "false"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.require", "any"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.within", "302"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.#", "2"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.expect.#", "3"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.expect.0", "prefect.flow-run.Cancelled"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.expect.1", "prefect.flow-run.Crashed"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.expect.2", "prefect.flow-run.Failed"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.threshold", "1"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.within", "0"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.after.#", "2"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.after.0", "prefect.flow-run.Completed"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.after.1", "prefect.flow-run.Succeeded"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.expect.#", "1"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.expect.0", "prefect.flow-run.Completed"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.threshold", "1"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.within", "0"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "actions.0.type", "run-deployment"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "actions.0.source", "inferred"),
					resource.TestCheckResourceAttr(compoundTriggerAutomationResourceNameAndPath, "actions.0.job_variables", testutils.NormalizedValueForJSON(t, `{"var1":"value1","var2":"value2","var3":"value3"}`)),
				),
			},
			{
				Config: fixtureAccAutomationResourceSequenceTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         sequenceTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutomationResourceExists(sequenceTriggerAutomationResourceNameAndPath, &api.Automation{}),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "name", "test-sequence-automation"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "description", "description for test-sequence-automation"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "enabled", "true"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.within", "180"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.#", "3"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.expect.#", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.expect.0", "prefect.flow-run.Pending"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.match_related", testutils.NormalizedValueForJSON(t, `{}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.threshold", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.0.event.within", "0"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.expect.#", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.expect.0", "prefect.flow-run.Running"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"],"prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.threshold", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.1.event.within", "0"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.expect.#", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.expect.0", "prefect.flow-run.Completed"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":["prefect.flow-run.*"],"prefect.resource.role":"flow"}`)),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.posture", "Reactive"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.threshold", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.triggers.2.event.within", "0"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "actions.#", "1"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "actions.0.type", "send-notification"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "actions.0.block_document_id", "123e4567-e89b-12d3-a456-426614174000"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "actions.0.subject", "Flow run failed"),
					resource.TestCheckResourceAttr(sequenceTriggerAutomationResourceNameAndPath, "actions.0.body", "Flow run failed at this time"),
				),
			},
		},
	})
}

func testAccCheckAutomationResourceExists(automationResourceName string, automation *api.Automation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		automationResource, ok := s.RootModule().Resources[automationResourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", automationResourceName)
		}

		automationID, _ := uuid.Parse(automationResource.Primary.ID)

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("workspace resource not found: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		automationClient, _ := c.Automations(uuid.Nil, workspaceID)
		fetchedAutomation, err := automationClient.Get(context.Background(), automationID)
		if err != nil {
			return fmt.Errorf("Error fetching Automation: %w", err)
		}
		if fetchedAutomation == nil {
			return fmt.Errorf("Automation not found for ID: %s", automationResource.Primary.ID)
		}

		*automation = *fetchedAutomation

		return nil
	}
}

func getAutomationImportStateID(automationResourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceResource, exists := state.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		automationResource, exists := state.RootModule().Resources[automationResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", automationResourceName)
		}
		automationID := automationResource.Primary.ID

		return fmt.Sprintf("%s,%s", automationID, workspaceID), nil
	}
}
