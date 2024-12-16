package resources_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccWebhook(workspace, name, template string, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "prefect_webhook" "%s" {
	name = "%s"
	template = jsonencode(%s)
	enabled = %t
	workspace_id = prefect_workspace.test.id
}
`, workspace, name, name, template, enabled)
}

const webhookTemplateDynamic = `
{
    "event": "model.refreshed",
    "resource": {
        "prefect.resource.id": "product.models.{{ body.model }}",
        "prefect.resource.name": "{{ body.friendly_name }}",
        "producing-team": "Data Science"
    }
}
`

const webhookTemplateStatic = `
{
    "event": "model.refreshed",
    "resource": {
        "prefect.resource.id": "product.models.recommendations",
        "prefect.resource.name": "Recommendations [Products]",
        "producing-team": "Data Science"
    }
}
`

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_webhook(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	randomName := testutils.NewRandomPrefixedString()
	webhookResourceName := "prefect_webhook." + randomName

	// We use this variable to store the fetched resource from the API
	// and it will be shared between TestSteps via a pointer.
	var webhook api.Webhook

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the work pool resource
				Config: fixtureAccWebhook(workspace.Resource, randomName, webhookTemplateDynamic, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(webhookResourceName, &webhook),
					resource.TestCheckResourceAttr(webhookResourceName, "name", randomName),
					resource.TestCheckResourceAttr(webhookResourceName, "enabled", "true"),
				),
			},
			{
				// Check that changing the paused state will update the resource in place
				Config: fixtureAccWebhook(workspace.Resource, randomName, webhookTemplateDynamic, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(webhookResourceName, &webhook),
					resource.TestCheckResourceAttr(webhookResourceName, "name", randomName),
					resource.TestCheckResourceAttr(webhookResourceName, "enabled", "false"),
				),
			},
			{
				// Check that changing the template will update the resource in place
				Config: fixtureAccWebhook(workspace.Resource, randomName, webhookTemplateStatic, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebhookExists(webhookResourceName, &webhook),
					resource.TestCheckResourceAttr(webhookResourceName, "name", randomName),
					resource.TestCheckResourceAttr(webhookResourceName, "enabled", "true"),
				),
			},
			// Import State checks - import by name (dynamic)
			{
				ImportState:       true,
				ResourceName:      webhookResourceName,
				ImportStateIdFunc: getWebhookImportStateID(webhookResourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckWebhookExists(webhookResourceName string, webhook *api.Webhook) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		webhookResource, exists := state.RootModule().Resources[webhookResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", webhookResourceName)
		}

		workspaceResource, exists := state.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		// Create a new client, and use the default configurations from the environment
		c, _ := testutils.NewTestClient()
		webhooksClient, _ := c.Webhooks(uuid.Nil, workspaceID)

		fetchedWebhook, err := webhooksClient.Get(context.Background(), webhookResource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error fetching webhook: %w", err)
		}
		if fetchedWebhook == nil {
			return fmt.Errorf("Webhook not found for ID: %s", webhookResource.Primary.ID)
		}

		*webhook = *fetchedWebhook

		return nil
	}
}

func getWebhookImportStateID(webhookResourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceResource, exists := state.RootModule().Resources[testutils.WorkspaceResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", testutils.WorkspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		webhookResource, exists := state.RootModule().Resources[webhookResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", webhookResourceName)
		}
		webhookID := webhookResource.Primary.ID

		return fmt.Sprintf("%s,%s", workspaceID, webhookID), nil
	}
}
