package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

type blockFixtureConfig struct {
	Workspace     string
	BlockName     string
	BlockValue    string
	RefBlockValue string
}

func fixtureAccBlock(cfg blockFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block" "{{ .BlockName }}" {
	name = "{{ .BlockName }}"
	type_slug = "secret"
	data = jsonencode({
		"value" = "{{ .BlockValue }}"
	})
	workspace_id = prefect_workspace.test.id
	depends_on = [prefect_workspace.test]
}`

	return testutils.RenderTemplate(tmpl, cfg)
}

func fixtureAccBlockWithRef(cfg blockFixtureConfig) string {
	tmpl := `
{{ .Workspace }}

resource "prefect_block" "{{ .BlockName }}" {
	name = "{{ .BlockName }}"
	type_slug = "secret"
	data = jsonencode({
		"value" = "{{ .BlockValue }}"
	})
	workspace_id = prefect_workspace.test.id
	depends_on = [prefect_workspace.test]
}

resource "prefect_block" "with_ref" {
  name      = "block-with-ref"
  type_slug = "s3-bucket"

  data = jsonencode({
    bucket_name = "my-bucket"
    credentials = { "$ref" : {{ .RefBlockValue }} }
  })

  workspace_id = prefect_workspace.test.id
  depends_on = [prefect_workspace.test]
}
`

	return testutils.RenderTemplate(tmpl, cfg)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block(t *testing.T) {
	randomName := testutils.NewRandomPrefixedString()
	randomValue := testutils.NewRandomPrefixedString()
	randomValue2 := testutils.NewRandomPrefixedString()

	workspace := testutils.NewEphemeralWorkspace()

	blockResourceName := fmt.Sprintf("prefect_block.%s", randomName)

	// We use this variable to store the fetched block document resource from the API
	// and it will be shared between the TestSteps via pointer.
	var blockDocument api.BlockDocument

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			// Check creation + existence of the block resource
			{
				Config: fixtureAccBlock(blockFixtureConfig{
					Workspace:  workspace.Resource,
					BlockName:  randomName,
					BlockValue: randomValue,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockExists(blockResourceName, &blockDocument),
					testAccCheckBlockValues(&blockDocument, ExpectedBlockValues{
						Name:     randomName,
						TypeSlug: "secret",
						Data:     fmt.Sprintf(`{"value":%q}`, randomValue),
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(blockResourceName, "name", randomName),
					testutils.ExpectKnownValue(blockResourceName, "type_slug", "secret"),
					testutils.ExpectKnownValue(blockResourceName, "data", fmt.Sprintf(`{"value":%q}`, randomValue)),
				},
			},
			// Check updating the value of the block resource
			{
				Config: fixtureAccBlock(blockFixtureConfig{
					Workspace:  workspace.Resource,
					BlockName:  randomName,
					BlockValue: randomValue2,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockExists(blockResourceName, &blockDocument),
					testAccCheckBlockValues(&blockDocument, ExpectedBlockValues{
						Name:     randomName,
						TypeSlug: "secret",
						Data:     fmt.Sprintf(`{"value":%q}`, randomValue2),
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(blockResourceName, "name", randomName),
					testutils.ExpectKnownValue(blockResourceName, "type_slug", "secret"),
					testutils.ExpectKnownValue(blockResourceName, "data", fmt.Sprintf(`{"value":%q}`, randomValue2)),
				},
			},
			// Next two tests using `fixtureAccBlockWithRef` will be used to test
			// that using the $ref syntax won't result in an Update plan if no changes are made.
			{
				Config: fixtureAccBlockWithRef(blockFixtureConfig{
					Workspace:     workspace.Resource,
					BlockName:     randomName,
					BlockValue:    randomValue2,
					RefBlockValue: fmt.Sprintf(`{"block_document_id":prefect_block.%s.id}`, randomName),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockExists("prefect_block.with_ref", &blockDocument),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue("prefect_block.with_ref", "name", "block-with-ref"),
					testutils.ExpectKnownValue("prefect_block.with_ref", "type_slug", "s3-bucket"),
				},
			},
			{
				Config: fixtureAccBlockWithRef(blockFixtureConfig{
					Workspace:     workspace.Resource,
					BlockName:     randomName,
					BlockValue:    randomValue2,
					RefBlockValue: fmt.Sprintf(`{"block_document_id":prefect_block.%s.id}`, randomName),
				}),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Import State checks - import by block_id,workspace_id (dynamic)
			// NOTE: the ImportStateVerify is set to false, as we need to omit the .Data
			// field when we hydrate the state from the API.
			{
				ImportState:       true,
				ResourceName:      blockResourceName,
				ImportStateIdFunc: testutils.GetResourceWorkspaceImportStateID(blockResourceName),
				ImportStateVerify: false,
			},
		},
	})
}

// testAccCheckBlockExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckBlockExists(blockResourceName string, blockDocument *api.BlockDocument) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the block resource we just created from the state
		blockID, err := testutils.GetResourceIDFromState(s, blockResourceName)
		if err != nil {
			return fmt.Errorf("error fetching block ID: %w", err)
		}

		// Get the workspace resource we just created from the state
		workspaceID, err := testutils.GetResourceWorkspaceIDFromState(s)
		if err != nil {
			return fmt.Errorf("error fetching workspace ID: %w", err)
		}

		// Initialize the client with the associated workspaceID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		blockDocumentsClient, _ := c.BlockDocuments(uuid.Nil, workspaceID)

		fetchedBlockDocument, err := blockDocumentsClient.Get(context.Background(), blockID)
		if err != nil {
			return fmt.Errorf("error fetching block document: %w", err)
		}

		// Assign the fetched block document to the passed pointer
		// so we can use it in the next test assertion
		*blockDocument = *fetchedBlockDocument

		return nil
	}
}

type ExpectedBlockValues struct {
	Name     string
	TypeSlug string
	Data     string
}

// testAccCheckBlockValues is a Custom Check Function that
// verifies that the API object matches the expected values.
func testAccCheckBlockValues(fetchedBlockDocument *api.BlockDocument, expectedValues ExpectedBlockValues) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedBlockDocument.Name != expectedValues.Name {
			return fmt.Errorf("Expected block name to be %s, got %s", expectedValues.Name, fetchedBlockDocument.Name)
		}
		if fetchedBlockDocument.BlockType.Slug != expectedValues.TypeSlug {
			return fmt.Errorf("Expected block type_slug to be %s, got %s", expectedValues.TypeSlug, fetchedBlockDocument.BlockType.Slug)
		}

		byteSlice, _ := json.Marshal(fetchedBlockDocument.Data)
		if string(byteSlice) != expectedValues.Data {
			return fmt.Errorf("Expected block data to be %s, got %s", expectedValues.Data, string(byteSlice))
		}

		return nil
	}
}
