package testutils

import "fmt"

// FixtureAccAutomationWorkspace creates a minimal block, flow, and deployment
// resource.
//
// It is used in acceptance tests that depend on these resources, like automations.
func FixtureAccAutomationDeployment(workspaceIDArg string) string {
	return fmt.Sprintf(`
resource "prefect_block" "test_block" {
	name = "test-block"
	type_slug = "github-repository"

	data = jsonencode({
		"repository_url": "https://github.com/foo/bar",
		"reference": "main"
	})

	%s
}

resource "prefect_flow" "test_flow" {
	name = "test-flow"

	%s
}

resource "prefect_deployment" "test_deployment" {
	name = "test-deployment"
	description = "Test description"
	flow_id = prefect_flow.test_flow.id
	storage_document_id = prefect_block.test_block.id

	depends_on = [prefect_flow.test_flow]

	%s
}
`, workspaceIDArg, workspaceIDArg, workspaceIDArg)
}
