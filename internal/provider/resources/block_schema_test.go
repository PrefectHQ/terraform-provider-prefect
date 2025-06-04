package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type blockSchemaFixtureConfig struct {
	Workspace       string
	WorkspaceIDArg  string
	BlockSchemaName string
	BlockTypeID     string
	BlockType       string
	Capabilities    []string
	Version         string
	Fields          string
	Checksum        string
}

func fixtureAccBlockSchema(cfg blockSchemaFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block_schema" "test" {
	{{ .WorkspaceIDArg }}

	block_type_id = "{{ .BlockTypeID }}"
	block_type = "{{ .BlockType }}"
	capabilities = [{{ range .Capabilities }}"{{ . }}", {{ end }}]
	version = "{{ .Version }}"
	fields = jsonencode({{ .Fields }})
	checksum = "{{ .Checksum }}"
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block_schema(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	resourceName := "prefect_block_schema.test"

	blockType := "string"
	blockTypeID := "" // need to get this from the `prefect_block_type` resource

	fields := `{"title": "x", "type": "object", "properties": {"foo": {"title": "Foo", "type": "string"}}, "required": ["foo"]}`

	expectedFields := testutils.NormalizedValueForJSON(t, fields)

	cfgCreate := blockSchemaFixtureConfig{
		Workspace:      workspace.Resource,
		WorkspaceIDArg: workspace.IDArg,
		BlockTypeID:    blockTypeID,
		BlockType:      blockType,
		Capabilities:   []string{"read", "write"},
		Version:        "1.0.0",
		Checksum:       "123",
		Fields:         fields,
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccBlockSchema(cfgCreate),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "block_type_id", blockTypeID),
					testutils.ExpectKnownValue(resourceName, "block_type", blockType),
					testutils.ExpectKnownValueList(resourceName, "capabilities", []string{"read", "write"}),
					testutils.ExpectKnownValue(resourceName, "version", "1.0.0"),
					testutils.ExpectKnownValue(resourceName, "fields", expectedFields),
					testutils.ExpectKnownValue(resourceName, "checksum", "123"),
				},
			},
		},
	})
}
