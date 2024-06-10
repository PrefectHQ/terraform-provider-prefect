package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccBlock(name string, value string) string {
	return fmt.Sprintf(`
resource "prefect_workspace" "workspace" {
	name = "%s"
	handle = "%s"
}
resource "prefect_block" "block" {
	name = "%s"
	type_slug = "secret"
	workspace_id = prefect_workspace.workspace.id
	data = jsonencode({
		"value": "%s"
	})
	depends_on = [prefect_workspace.workspace]
}
	`, name, name, name, value)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_block(t *testing.T) {
	blockResourceName := "prefect_block.block"
	const workspaceResourceName = "prefect_workspace.workspace"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomValue := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomValue2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// We use this variable to store the fetched block document resource from the API
	// and it will be shared between the TestSteps via pointer.
	var blockDocument api.BlockDocument

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			// Check creation + existence of the block resource
			{
				Config: fixtureAccBlock(randomName, randomValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockExists(blockResourceName, workspaceResourceName, &blockDocument),
					testAccCheckBlockValues(&blockDocument, ExpectedBlockValues{
						Name:     randomName,
						TypeSlug: "secret",
						Data:     fmt.Sprintf(`{"value":%q}`, randomValue),
					}),
					resource.TestCheckResourceAttr(blockResourceName, "name", randomName),
					resource.TestCheckResourceAttr(blockResourceName, "type_slug", "secret"),
					resource.TestCheckResourceAttr(blockResourceName, "data", fmt.Sprintf(`{"value":%q}`, randomValue)),
				),
			},
			// Check updating the value of the block resource
			{
				Config: fixtureAccBlock(randomName, randomValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBlockExists(blockResourceName, workspaceResourceName, &blockDocument),
					testAccCheckBlockValues(&blockDocument, ExpectedBlockValues{
						Name:     randomName,
						TypeSlug: "secret",
						Data:     fmt.Sprintf(`{"value":%q}`, randomValue2),
					}),
					resource.TestCheckResourceAttr(blockResourceName, "name", randomName),
					resource.TestCheckResourceAttr(blockResourceName, "type_slug", "secret"),
					resource.TestCheckResourceAttr(blockResourceName, "data", fmt.Sprintf(`{"value":%q}`, randomValue2)),
				),
			},
			// Import State checks - import by block_id,workspace_id (dynamic)
			{
				ImportState:       true,
				ResourceName:      blockResourceName,
				ImportStateIdFunc: getBlockImportStateID(blockResourceName, workspaceResourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// testAccCheckBlockExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckBlockExists(blockResourceName string, workspaceResourceName string, blockDocument *api.BlockDocument) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the block resource we just created from the state
		blockResource, exists := s.RootModule().Resources[blockResourceName]
		if !exists {
			return fmt.Errorf("resource not found: %s", blockResourceName)
		}
		blockID, _ := uuid.Parse(blockResource.Primary.ID)

		// Get the workspace resource we just created from the state
		workspaceResource, exists := s.RootModule().Resources[workspaceResourceName]
		if !exists {
			return fmt.Errorf("resource not found: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

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

// getBlockImportStateID generates the Import ID used in the test assertion,
// since we need to construct one that includes the Block ID and the Workspace ID.
func getBlockImportStateID(blockResourceName string, workspaceResourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceResource, exists := state.RootModule().Resources[workspaceResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", workspaceResourceName)
		}
		workspaceID, _ := uuid.Parse(workspaceResource.Primary.ID)

		blockResource, exists := state.RootModule().Resources[blockResourceName]
		if !exists {
			return "", fmt.Errorf("Resource not found in state: %s", blockResourceName)
		}
		blockID, _ := uuid.Parse(blockResource.Primary.ID)

		return fmt.Sprintf("%s,%s", blockID, workspaceID), nil
	}
}
