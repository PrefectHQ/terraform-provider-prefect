package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type slaResourceConfig struct {
	WorkspaceResource     string
	WorkspaceResourceName string
}

func fixtureAccSLAResourceCreate(cfg slaResourceConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "foo" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_deployment" "bar" {
	name = "my-deployment"
	flow_id = prefect_flow.foo.id
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_resource_sla" "test" {
	resource_id = "prefect.deployment.${prefect_deployment.bar.id}"
	workspace_id = {{.WorkspaceResourceName}}.id
	slas = [
		{
			name = "my-time-to-completion-sla"
			severity = "high"
			duration = 60
		},
		{
			name = "my-frequency-sla"
			severity = "critical"
			stale_after = 120
		},
		{
			name = "my-lateness-sla"
			severity = "moderate"
			within = 55
		},
		{
			name = "my-freshness-sla"
			severity = "moderate"
			within = 360
			resource_match = jsonencode({
				label = "my-label"
			})
			expected_event = "my-event"
		},
	]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccSLAResourceUpdate(cfg slaResourceConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_flow" "foo" {
	name = "my-flow"
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_deployment" "bar" {
	name = "my-deployment"
	flow_id = prefect_flow.foo.id
	workspace_id = {{.WorkspaceResourceName}}.id
}

resource "prefect_resource_sla" "test" {
	resource_id = "prefect.deployment.${prefect_deployment.bar.id}"
	workspace_id = {{.WorkspaceResourceName}}.id
	slas = [
		{
			name = "my-time-to-completion-sla"
			duration = 100
		},
		{
			name = "my-lateness-sla"
			within = 500
		},
		{
			name = "my-freshness-sla"
			within = 30
			resource_match = jsonencode({
				label = "my-label"
			})
			expected_event = "my-other-event"
		},
	]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_resource_sla(t *testing.T) {
	// SLAs are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	resourceName := "prefect_resource_sla.test"
	workspace := testutils.NewEphemeralWorkspace()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccSLAResourceCreate(slaResourceConfig{
					WorkspaceResource:     workspace.Resource,
					WorkspaceResourceName: testutils.WorkspaceResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueListSize(resourceName, "slas", 4),
					testutils.ExpectKnownValue(resourceName, "slas.0.name", "my-time-to-completion-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.0.duration", 60),
					testutils.ExpectKnownValue(resourceName, "slas.1.name", "my-frequency-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.1.stale_after", 120),
					testutils.ExpectKnownValue(resourceName, "slas.2.name", "my-lateness-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.2.within", 55),
					testutils.ExpectKnownValue(resourceName, "slas.3.name", "my-freshness-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.3.within", 360),
					testutils.ExpectKnownValue(resourceName, "slas.3.resource_match", `{"label":"my-label"}`),
					testutils.ExpectKnownValue(resourceName, "slas.3.expected_event", "my-event"),
				},
			},
			{
				Config: fixtureAccSLAResourceUpdate(slaResourceConfig{
					WorkspaceResource:     workspace.Resource,
					WorkspaceResourceName: testutils.WorkspaceResourceName,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueListSize(resourceName, "slas", 3),
					testutils.ExpectKnownValue(resourceName, "slas.0.name", "my-time-to-completion-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.0.duration", 100),
					testutils.ExpectKnownValue(resourceName, "slas.1.name", "my-lateness-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.1.within", 500),
					testutils.ExpectKnownValue(resourceName, "slas.2.name", "my-freshness-sla"),
					testutils.ExpectKnownValueNumber(resourceName, "slas.2.within", 30),
					testutils.ExpectKnownValue(resourceName, "slas.2.resource_match", `{"label":"my-label"}`),
					testutils.ExpectKnownValue(resourceName, "slas.2.expected_event", "my-other-event"),
				},
			},
		},
	})
}
