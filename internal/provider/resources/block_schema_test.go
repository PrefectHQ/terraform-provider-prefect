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

	BlockTypeName string
	BlockTypeSlug string

	Capabilities []string
	Version      string
	Fields       string
	Checksum     string
}

func fixtureAccBlockSchema(cfg blockSchemaFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block_type" "test" {
	name = "{{ .BlockTypeName }}"
	slug = "{{ .BlockTypeSlug }}"

	logo_url = "https://example.com/logo.png"
	documentation_url = "https://example.com/documentation"
	description = "test"
	code_example = "test"

	{{ .WorkspaceIDArg }}
}

resource "prefect_block_schema" "test" {
	{{ .WorkspaceIDArg }}

	block_type_id = prefect_block_type.test.id
	capabilities = [{{ range .Capabilities }}"{{ . }}", {{ end }}]
	version = "{{ .Version }}"
	fields = jsonencode({{ .Fields }})
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block_schema(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	resourceName := "prefect_block_schema.test"

	blockTypeName := testutils.NewRandomPrefixedString()

	fields := `{"title": "x", "type": "object", "properties": {"foo": {"title": "Foo", "type": "string"}}, "required": ["foo"]}`
	expectedFields := testutils.NormalizedValueForJSON(t, fields)

	fieldsUpdate := `{"title": "y", "type": "object", "properties": {"bar": {"title": "Bar", "type": "string"}}, "required": ["bar"]}`
	expectedFieldsUpdate := testutils.NormalizedValueForJSON(t, fieldsUpdate)

	cfgCreate := blockSchemaFixtureConfig{
		Workspace:      workspace.Resource,
		WorkspaceIDArg: workspace.IDArg,

		BlockTypeName: blockTypeName,
		BlockTypeSlug: blockTypeName,

		Capabilities: []string{"read", "write"},
		Version:      "1.0.0",
		Fields:       fields,
	}

	// Update is not actually implemented, so we're testing that it forces a recreation.
	cfgUpdate := blockSchemaFixtureConfig{
		// Reuse the workspace.
		Workspace:      cfgCreate.Workspace,
		WorkspaceIDArg: cfgCreate.WorkspaceIDArg,

		// Reuse fields for the block type resource.
		BlockTypeName: cfgCreate.BlockTypeName,
		BlockTypeSlug: cfgCreate.BlockTypeSlug,

		// Update values that can be updated.
		Capabilities: []string{"read", "write", "delete"},
		Version:      "1.0.1",
		Fields:       fieldsUpdate,
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccBlockSchema(cfgCreate),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check required/optional fields.
					testutils.ExpectKnownValueNotNull(resourceName, "block_type_id"),
					testutils.ExpectKnownValueList(resourceName, "capabilities", cfgCreate.Capabilities),
					testutils.ExpectKnownValue(resourceName, "version", cfgCreate.Version),
					testutils.ExpectKnownValue(resourceName, "fields", expectedFields),

					// Check computed fields.
					testutils.ExpectKnownValueNotNull(resourceName, "block_type"),
					testutils.ExpectKnownValueNotNull(resourceName, "checksum"),
				},
			},
			{
				Config: fixtureAccBlockSchema(cfgUpdate),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check required/optional fields.
					testutils.ExpectKnownValueNotNull(resourceName, "block_type_id"),
					testutils.ExpectKnownValueList(resourceName, "capabilities", cfgUpdate.Capabilities),
					testutils.ExpectKnownValue(resourceName, "version", cfgUpdate.Version),
					testutils.ExpectKnownValue(resourceName, "fields", expectedFieldsUpdate),

					// Check computed fields.
					testutils.ExpectKnownValueNotNull(resourceName, "block_type"),
					testutils.ExpectKnownValueNotNull(resourceName, "checksum"),
				},
			},
		},
	})
}
