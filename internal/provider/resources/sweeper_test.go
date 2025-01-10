package resources_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func addWorkspaceSweeper(testSuffix string, workspaces []string) {
	sweeperName := fmt.Sprintf("workspaces_%s", testSuffix)

	resource.AddTestSweepers(sweeperName, &resource.Sweeper{
		Name: sweeperName,
		F: func(_ string) error {
			ctx := context.Background()
			client, _ := testutils.NewTestClient()

			// NOTE: the accountID is inherited by the one set in the test environment
			workspacesClient, _ := client.Workspaces(uuid.Nil)

			workspaces, err := workspacesClient.List(ctx, workspaces)
			if err != nil {
				log.Printf("unable to list workspaces: %v", err)
			}

			for _, workspace := range workspaces {
				log.Printf("found workspace: %s", workspace.Name)
				if strings.HasPrefix(workspace.Name, testutils.TestAccPrefix) {
					log.Printf("found matching workspace %s, deleting...", workspace.Name)
					err := workspacesClient.Delete(context.Background(), workspace.ID)

					if err != nil {
						log.Printf("Error destroying %s during sweep: %s", workspace.Name, err)
					}
				} else {
					log.Printf("workspace %s did not match prefix, skipping...", workspace.Name)
				}
			}

			return nil
		},
	})
}
