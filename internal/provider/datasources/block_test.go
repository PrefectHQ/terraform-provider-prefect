package datasources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type blockFixtureConfig struct {
	AccountID string
	Workspace string
	BlockName string
}

func fixtureAccBlock(cfg blockFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block" "{{ .BlockName }}" {
  name      = "{{ .BlockName }}"
  type_slug = "secret"

  data = jsonencode({
    "someKey" = "someValue"
  })

  account_id = "{{ .AccountID }}"
  workspace_id = prefect_workspace.test.id
}

data "prefect_block" "my_existing_secret_by_id" {
  id = prefect_block.{{ .BlockName }}.id

  account_id = "{{ .AccountID }}"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_block.{{ .BlockName }}]
}

data "prefect_block" "my_existing_secret_by_name" {
  name      = "{{ .BlockName }}"
  type_slug = "secret"

  account_id = "{{ .AccountID }}"
  workspace_id = prefect_workspace.test.id

  depends_on = [prefect_block.{{ .BlockName }}]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_block(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()

	datasourceNameByID := "data.prefect_block.my_existing_secret_by_id"
	datasourceNameByName := "data.prefect_block.my_existing_secret_by_name"

	blockName := "my-block"
	blockValue := "{\"someKey\":\"someValue\"}"

	cfg := blockFixtureConfig{
		Workspace: workspace.Resource,
		BlockName: blockName,
		AccountID: os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"),
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test block datasource by ID.
				Config: fixtureAccBlock(cfg),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByID, "id"),
					testutils.ExpectKnownValue(datasourceNameByID, "name", blockName),
					testutils.ExpectKnownValue(datasourceNameByID, "data", blockValue),
				},
			},
			{
				// Test block datasource by name.
				Config: fixtureAccBlock(cfg),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueNotNull(datasourceNameByName, "id"),
					testutils.ExpectKnownValue(datasourceNameByName, "name", blockName),
					testutils.ExpectKnownValue(datasourceNameByName, "data", blockValue),
				},
			},
		},
	})
}
