package datasources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type deploymentFixtureConfig struct {
	Workspace      string
	WorkspaceIDArg string
	AccountID      string
}

func fixtureAccDeployment(cfg deploymentFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_flow" "test" {
	name = "test"
	tags = ["test"]

	{{.WorkspaceIDArg}}
}

resource "prefect_deployment" "test" {
  name = "test"
  flow_id = prefect_flow.test.id

  description = "test description"
  version = "1.2.3"

	{{.WorkspaceIDArg}}
}

data "prefect_deployment" "test_by_id" {
  id = prefect_deployment.test.id

	{{.WorkspaceIDArg}}

  depends_on = [prefect_deployment.test]
}

data "prefect_deployment" "test_by_name" {
  name = prefect_deployment.test.name
  flow_name = prefect_flow.test.name

	{{.WorkspaceIDArg}}

  depends_on = [prefect_deployment.test]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_deployment(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	datasourceNameByID := "data.prefect_deployment.test_by_id"
	datasourceNameByName := "data.prefect_deployment.test_by_name"

	cfg := deploymentFixtureConfig{
		Workspace:      workspace.Resource,
		WorkspaceIDArg: workspace.IDArg,
		AccountID:      os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"),
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccDeployment(cfg),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByID, "id"),
					testutils.ExpectKnownValue(datasourceNameByID, "name", "test"),
					testutils.ExpectKnownValue(datasourceNameByID, "description", "test description"),
					testutils.ExpectKnownValue(datasourceNameByID, "version", "1.2.3"),
				},
			},
			{
				Config: fixtureAccDeployment(cfg),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByName, "id"),
					testutils.ExpectKnownValue(datasourceNameByName, "name", "test"),
					testutils.ExpectKnownValue(datasourceNameByName, "description", "test description"),
					testutils.ExpectKnownValue(datasourceNameByName, "version", "1.2.3"),
				},
			},
		},
	})
}
