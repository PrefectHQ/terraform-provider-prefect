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

// TestMain adds sweeper functionality to the "go test" command.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

//nolint:gochecknoinits // required per Terraform testing guide
func init() {
	addWorkspaceSweeper()
}

// addWorkspaceSweeper adds a sweeper that deletes any workspaces that match
// the prefix we use for ephemeral workspaces in acceptance tests.
//
// This is designed to run at a given interval when other acceptance tests are
// not likely running.
func addWorkspaceSweeper() {
	resource.AddTestSweepers("workspaces", &resource.Sweeper{
		Name: "workspaces",
		F: func(_ string) error {
			client, err := testutils.NewTestClient()
			if err != nil {
				return fmt.Errorf("unable to get prefect client: %w", err)
			}

			// NOTE: the accountID is inherited by the one set in the test environment
			workspacesClient, err := client.Workspaces(uuid.Nil)
			if err != nil {
				return fmt.Errorf("unable to get workspaces client: %w", err)
			}

			workspaces, err := workspacesClient.List(context.Background(), []string{})
			if err != nil {
				return fmt.Errorf("unable to list workspaces: %w", err)
			}

			if len(workspaces) == 0 {
				return fmt.Errorf("no workspaces found for this account")
			}

			for _, workspace := range workspaces {
				if strings.HasPrefix(workspace.Name, testutils.TestAccPrefix) {
					log.Printf("found acceptance testing workspace %s, deleting...\n", workspace.Name)

					err := workspacesClient.Delete(context.Background(), workspace.ID)
					if err != nil {
						log.Printf("unable to delete workspaces %s during sweep: %s\n", workspace.Name, err)
					}
				} else {
					log.Printf("workspace %s does not match acceptance testing prefix, skipping...\n", workspace.Name)
				}
			}

			return nil
		},
	})
}
