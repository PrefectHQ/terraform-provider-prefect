package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type blockTypeFixtureConfig struct {
	Workspace      string
	WorkspaceIDArg string

	Name string
	Slug string

	LogoURL          string
	DocumentationURL string
	Description      string
	CodeExample      string
}

func fixtureAccBlockType(cfg blockTypeFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block_type" "test" {
	name = "{{ .Name }}"
	slug = "{{ .Slug }}"

	logo_url = "{{ .LogoURL }}"
	documentation_url = "{{ .DocumentationURL }}"
	description = "{{ .Description }}"
	code_example = "{{ .CodeExample }}"

	{{ .WorkspaceIDArg }}
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block_type(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	blockTypeResourceName := "prefect_block_type.test"

	cfgCreate := blockTypeFixtureConfig{
		Workspace:      workspace.Resource,
		WorkspaceIDArg: workspace.IDArg,

		Name:             "test",
		Slug:             "test",
		LogoURL:          "https://example.com/logo.png",
		DocumentationURL: "https://example.com/documentation",
		Description:      "test",
		CodeExample:      "test",
	}

	cfgUpdate := blockTypeFixtureConfig{
		// Use the same workspace.
		Workspace:      cfgCreate.Workspace,
		WorkspaceIDArg: cfgCreate.WorkspaceIDArg,

		// Reuse values that cannot be updated.
		Name: cfgCreate.Name,
		Slug: cfgCreate.Slug,

		// Update values that can be updated.
		LogoURL:          "https://example.com/logo-updated.png",
		DocumentationURL: "https://example.com/documentation-updated",
		Description:      "test-updated",
		CodeExample:      "test-updated",
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccBlockType(cfgCreate),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(blockTypeResourceName, "name", cfgCreate.Name),
					testutils.ExpectKnownValue(blockTypeResourceName, "slug", cfgCreate.Slug),
					testutils.ExpectKnownValue(blockTypeResourceName, "logo_url", cfgCreate.LogoURL),
					testutils.ExpectKnownValue(blockTypeResourceName, "documentation_url", cfgCreate.DocumentationURL),
					testutils.ExpectKnownValue(blockTypeResourceName, "description", cfgCreate.Description),
					testutils.ExpectKnownValue(blockTypeResourceName, "code_example", cfgCreate.CodeExample),
					testutils.ExpectKnownValueBool(blockTypeResourceName, "is_protected", false),
				},
			},
			{
				Config: fixtureAccBlockType(cfgUpdate),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(blockTypeResourceName, "name", cfgUpdate.Name),
					testutils.ExpectKnownValue(blockTypeResourceName, "slug", cfgUpdate.Slug),
					testutils.ExpectKnownValue(blockTypeResourceName, "logo_url", cfgUpdate.LogoURL),
					testutils.ExpectKnownValue(blockTypeResourceName, "documentation_url", cfgUpdate.DocumentationURL),
					testutils.ExpectKnownValue(blockTypeResourceName, "description", cfgUpdate.Description),
					testutils.ExpectKnownValue(blockTypeResourceName, "code_example", cfgUpdate.CodeExample),
					testutils.ExpectKnownValueBool(blockTypeResourceName, "is_protected", false),
				},
			},
		},
	})
}
