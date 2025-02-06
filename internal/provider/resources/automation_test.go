package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "name", "test-event-automation"),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "description", "description for test-event-automation"),
					testutils.ExpectKnownValueBool(eventTriggerAutomationResourceNameAndPath, "enabled", true),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "trigger.event.posture", "Reactive"),
					testutils.ExpectKnownValueList(eventTriggerAutomationResourceNameAndPath, "trigger.event.after", []string{"prefect.flow-run.Completed", "prefect.flow-run.Succeeded"}),
					testutils.ExpectKnownValueList(eventTriggerAutomationResourceNameAndPath, "trigger.event.expect", []string{"prefect.flow-run.Cancelled", "prefect.flow-run.Crashed", "prefect.flow-run.Failed"}),
					testutils.ExpectKnownValueList(eventTriggerAutomationResourceNameAndPath, "trigger.event.for_each", []string{"prefect.resource.id", "prefect.resource.role"}),
					testutils.ExpectKnownValueNumber(eventTriggerAutomationResourceNameAndPath, "trigger.event.threshold", 1),
					testutils.ExpectKnownValueNumber(eventTriggerAutomationResourceNameAndPath, "trigger.event.within", 60),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "actions.0.type", "run-deployment"),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "actions.0.source", "selected"),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "actions.0.deployment_id", "123e4567-e89b-12d3-a456-426614174000"),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "trigger.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "trigger.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"],"prefect.resource.role":"flow"}`)),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "actions.0.parameters", testutils.NormalizedValueForJSON(t, `{"param1":"value1","param2":"value2"}`)),
					testutils.ExpectKnownValue(eventTriggerAutomationResourceNameAndPath, "actions.0.job_variables", testutils.NormalizedValueForJSON(t, `{"string_var":"value1","int_var":2,"bool_var":true}`)),
				},
			},
			// Import State checks - import by automation_id
			{
				ImportState:       true,
				ResourceName:      eventTriggerAutomationResourceNameAndPath,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(eventTriggerAutomationResourceNameAndPath),
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "name", "test-metric-automation"),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "description", "description for test-metric-automation"),
					testutils.ExpectKnownValueBool(metricTriggerAutomationResourceNameAndPath, "enabled", true),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "trigger.metric.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "trigger.metric.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.name", "duration"),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.operator", ">="),
					testutils.ExpectKnownValueFloat(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.threshold", 0.5),
					testutils.ExpectKnownValueNumber(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.range", 30),
					testutils.ExpectKnownValueNumber(metricTriggerAutomationResourceNameAndPath, "trigger.metric.metric.firing_for", 60),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "actions.0.type", "change-flow-run-state"),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "actions.0.state", "FAILED"),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "actions.0.name", "Failed by automation"),
					testutils.ExpectKnownValue(metricTriggerAutomationResourceNameAndPath, "actions.0.message", "Flow run failed"),
				},
			},
			// Import State checks - import by automation_id
			{
				ImportState:       true,
				ResourceName:      metricTriggerAutomationResourceNameAndPath,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(metricTriggerAutomationResourceNameAndPath),
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "name", "test-compound-automation"),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "description", "description for test-compound-automation"),
					testutils.ExpectKnownValueBool(compoundTriggerAutomationResourceNameAndPath, "enabled", false),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.require", "any"),
					testutils.ExpectKnownValueNumber(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.within", 302),
					testutils.ExpectKnownValueList(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.expect", []string{
						"prefect.flow-run.Cancelled",
						"prefect.flow-run.Crashed",
						"prefect.flow-run.Failed",
					}),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.posture", "Reactive"),
					testutils.ExpectKnownValueNumber(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.threshold", 1),
					testutils.ExpectKnownValueNumber(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.within", 0),
					testutils.ExpectKnownValueList(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.0.event.after", []string{
						"prefect.flow-run.Completed",
						"prefect.flow-run.Succeeded",
					}),
					testutils.ExpectKnownValueList(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.expect", []string{
						"prefect.flow-run.Completed",
					}),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.match", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*"}`)),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.match_related", testutils.NormalizedValueForJSON(t, `{"prefect.resource.id":"prefect.flow-run.*","prefect.resource.role":"flow"}`)),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.posture", "Reactive"),
					testutils.ExpectKnownValueNumber(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.threshold", 1),
					testutils.ExpectKnownValueNumber(compoundTriggerAutomationResourceNameAndPath, "trigger.compound.triggers.1.event.within", 0),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "actions.0.type", "run-deployment"),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "actions.0.source", "inferred"),
					testutils.ExpectKnownValue(compoundTriggerAutomationResourceNameAndPath, "actions.0.job_variables", testutils.NormalizedValueForJSON(t, `{"var1":"value1","var2":"value2","var3":"value3"}`)),
				},
			},
			{
				Config: fixtureAccAutomationResourceSequenceTrigger(automationFixtureConfig{
					EphemeralWorkspace:             ephemeralWorkspace.Resource,
					EphemeralWorkspaceResourceName: testutils.WorkspaceResourceName,
					AutomationResourceName:         sequenceTriggerAutomationResourceName,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutomationResourceExists(sequenceTriggerAutomationResourceNameAndPath, &api.Automation{}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath, "name", "test-sequence-automation"),
					testutils.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath, "description", "description for test-sequence-automation"),
					testutils.ExpectKnownValueBool(sequenceTriggerAutomationResourceNameAndPath, "enabled", true),
					testutils.ExpectKnownValueNumber(sequenceTriggerAutomationResourceNameAndPath, "trigger.sequence.within", 180),

					statecheck.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath,
						tfjsonpath.New("trigger").AtMapKey("sequence").AtMapKey("triggers").AtSliceIndex(0).AtMapKey("event"),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"expect":        knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("prefect.flow-run.Pending")}),
							"match":         knownvalue.StringExact(`{"prefect.resource.id":"prefect.flow-run.*"}`),
							"match_related": knownvalue.StringExact(`{}`),
							"posture":       knownvalue.StringExact("Reactive"),
							"threshold":     knownvalue.Int64Exact(1),
							"within":        knownvalue.Int64Exact(0),
						}),
					),

					statecheck.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath,
						tfjsonpath.New("trigger").AtMapKey("sequence").AtMapKey("triggers").AtSliceIndex(1).AtMapKey("event"),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"expect":        knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("prefect.flow-run.Running")}),
							"match":         knownvalue.StringExact(`{"prefect.resource.id":"prefect.flow-run.*"}`),
							"match_related": knownvalue.StringExact(`{"prefect.resource.id":["prefect.flow.ce6ec0c9-4b51-483b-a776-43c085b6c4f8"],"prefect.resource.role":"flow"}`),
							"posture":       knownvalue.StringExact("Reactive"),
							"threshold":     knownvalue.Int64Exact(1),
							"within":        knownvalue.Int64Exact(0),
						}),
					),

					statecheck.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath,
						tfjsonpath.New("trigger").AtMapKey("sequence").AtMapKey("triggers").AtSliceIndex(2).AtMapKey("event"),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"expect":        knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("prefect.flow-run.Completed")}),
							"match":         knownvalue.StringExact(`{"prefect.resource.id":"prefect.flow-run.*"}`),
							"match_related": knownvalue.StringExact(`{"prefect.resource.id":["prefect.flow-run.*"],"prefect.resource.role":"flow"}`),
							"posture":       knownvalue.StringExact("Reactive"),
							"threshold":     knownvalue.Int64Exact(1),
							"within":        knownvalue.Int64Exact(0),
						}),
					),

					statecheck.ExpectKnownValue(sequenceTriggerAutomationResourceNameAndPath,
						tfjsonpath.New("actions").AtSliceIndex(0),
						knownvalue.MapPartial(map[string]knownvalue.Check{
							"type":              knownvalue.StringExact("send-notification"),
							"block_document_id": knownvalue.StringExact("123e4567-e89b-12d3-a456-426614174000"),
							"subject":           knownvalue.StringExact("Flow run failed"),
							"body":              knownvalue.StringExact("Flow run failed at this time"),
						}),
					),
				},
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
