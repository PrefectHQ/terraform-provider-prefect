package testutils

import "fmt"

// FixtureAccAutomationWorkspace creates a minimal block, flow, and deployment
// resource.
//
// It is used in acceptance tests that depend on these resources, like automations.
// The block, flow, and deployment names are randomized so that the resource and
// datasource automation tests (which both render this fixture) do not collide on
// a shared name when they run in parallel against the same instance.
func FixtureAccAutomationDeployment(workspaceIDArg string) string {
	blockName := NewRandomPrefixedString()
	flowName := NewRandomPrefixedString()
	deploymentName := NewRandomPrefixedString()

	return fmt.Sprintf(`
resource "prefect_block" "test_block" {
	name = "%s"
	type_slug = "secret"

	data = jsonencode({
		"value": "test-value"
	})

	%s
}

resource "prefect_flow" "test_flow" {
	name = "%s"

	%s
}

resource "prefect_deployment" "test_deployment" {
	name = "%s"
	description = "Test description"
	flow_id = prefect_flow.test_flow.id
	storage_document_id = prefect_block.test_block.id

	depends_on = [prefect_flow.test_flow]

	%s
}
`, blockName, workspaceIDArg, flowName, workspaceIDArg, deploymentName, workspaceIDArg)
}
