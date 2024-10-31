package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentAccessConfig struct {
	FlowName              string
	DeploymentName        string
	ServiceAccountName    string
	WorkspaceResource     string
	WorkspaceResourceName string
}

func fixtureAccDeploymentAccess(cfg deploymentAccessConfig) string {
	tmpl := `
{{.WorkspaceResource}}

resource "prefect_service_account" "test" {
	name = "{{.ServiceAccountName}}"
}

data "prefect_team" "test" {
	name = "my-team"
}

resource "prefect_flow" "test" {
	name = "{{.FlowName}}"
	workspace_id = {{.WorkspaceResourceName}}.id
	tags = ["test"]
}

resource "prefect_deployment" "test" {
	name = "{{.DeploymentName}}"
	workspace_id = {{.WorkspaceResourceName}}.id
	flow_id = prefect_flow.test.id
}

resource "prefect_deployment_access" "test" {
	deployment_id = prefect_deployment.test.id

	workspace_id = {{.WorkspaceResourceName}}.id

	manage_actor_ids = [prefect_service_account.test.id]
	run_actor_ids = [prefect_service_account.test.id]
	view_actor_ids = [prefect_service_account.test.id]
	manage_team_ids = [data.prefect_team.test.id]
	run_team_ids = [data.prefect_team.test.id]
	view_team_ids = [data.prefect_team.test.id]
}
`

	return helpers.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_deployment_access(t *testing.T) {
	deploymentName := testutils.NewRandomPrefixedString()
	serviceAccountName := testutils.NewRandomPrefixedString()
	flowName := testutils.NewRandomPrefixedString()

	workspace, workspaceName := testutils.NewEphemeralWorkspace()
	workspaceResourceName := "prefect_workspace." + workspaceName

	cfgSet := deploymentAccessConfig{
		FlowName:              flowName,
		DeploymentName:        deploymentName,
		ServiceAccountName:    serviceAccountName,
		WorkspaceResource:     workspace,
		WorkspaceResourceName: workspaceResourceName,
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeploymentAccess(cfgSet),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"prefect_deployment_access.test", "deployment_id",
						"prefect_deployment.test", "id",
					),
				),
			},
		},
	})
}
