package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

// AddWorkspaceSweeper adds a sweeper that deletes any workspaces that match
// the prefix we use for ephemeral workspaces in acceptance tests.
//
// This is designed to run at a given interval when other acceptance tests are
// not likely running.
func AddWorkspaceSweeper() {
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

// AddServiceAccountSweeper adds a sweeper that deletes any service accounts that match
// the prefix we use for ephemeral service accounts in acceptance tests.
//
// This is designed to run at a given interval when other acceptance tests are
// not likely running.
func AddServiceAccountSweeper() {
	resource.AddTestSweepers("serviceAccounts", &resource.Sweeper{
		Name: "serviceAccounts",
		F: func(_ string) error {
			client, err := testutils.NewTestClient()
			if err != nil {
				return fmt.Errorf("unable to get prefect client: %w", err)
			}

			// NOTE: the accountID is inherited by the one set in the test environment
			serviceAccountsClient, err := client.ServiceAccounts(uuid.Nil)
			if err != nil {
				return fmt.Errorf("unable to get service accounts client: %w", err)
			}

			serviceAccounts, err := serviceAccountsClient.List(context.Background(), []string{})
			if err != nil {
				return fmt.Errorf("unable to list service accounts: %w", err)
			}

			if len(serviceAccounts) == 0 {
				return fmt.Errorf("no service accounts found for this account")
			}

			for _, serviceAccount := range serviceAccounts {
				if strings.HasPrefix(serviceAccount.Name, testutils.TestAccPrefix) {
					log.Printf("found acceptance testing service account %s, deleting...\n", serviceAccount.Name)

					err := serviceAccountsClient.Delete(context.Background(), serviceAccount.ID.String())
					if err != nil {
						log.Printf("unable to delete service accounts %s during sweep: %s\n", serviceAccount.Name, err)
					}
				} else {
					log.Printf("service account %s does not match acceptance testing prefix, skipping...\n", serviceAccount.Name)
				}
			}

			return nil
		},
	})
}

// AddWorkspaceRoleSweeper adds a sweeper that deletes any workspace roles that match
// the prefix we use for ephemeral workspace roles in acceptance tests.
//
// This is designed to run at a given interval when other acceptance tests are
// not likely running.
func AddWorkspaceRoleSweeper() {
	resource.AddTestSweepers("WorkspaceRoles", &resource.Sweeper{
		Name: "workspaceRoles",
		F: func(_ string) error {
			client, err := testutils.NewTestClient()
			if err != nil {
				return fmt.Errorf("unable to get prefect client: %w", err)
			}

			// NOTE: the workspaceID is inherited by the one set in the test environment
			workspaceRolesClient, err := client.WorkspaceRoles(uuid.Nil)
			if err != nil {
				return fmt.Errorf("unable to get workspace roles client: %w", err)
			}

			workspaceRoles, err := workspaceRolesClient.List(context.Background(), []string{})
			if err != nil {
				return fmt.Errorf("unable to list workspace roles: %w", err)
			}

			if len(workspaceRoles) == 0 {
				return fmt.Errorf("no workspace roles found for this workspace")
			}

			for _, workspaceRole := range workspaceRoles {
				if strings.HasPrefix(workspaceRole.Name, testutils.TestAccPrefix) {
					log.Printf("found acceptance testing workspace role %s, deleting...\n", workspaceRole.Name)

					err := workspaceRolesClient.Delete(context.Background(), workspaceRole.ID)
					if err != nil {
						log.Printf("unable to delete workspace roles %s during sweep: %s\n", workspaceRole.Name, err)
					}
				} else {
					log.Printf("workspace role %s does not match acceptance testing prefix, skipping...\n", workspaceRole.Name)
				}
			}

			return nil
		},
	})
}

// AddTeamSweeper adds a sweeper that deletes any teams that match
// the prefix we use for ephemeral teams in acceptance tests.
//
// This is designed to run at a given interval when other acceptance tests are
// not likely running.
func AddTeamSweeper() {
	resource.AddTestSweepers("teams", &resource.Sweeper{
		Name: "teams",
		F: func(_ string) error {
			client, err := testutils.NewTestClient()
			if err != nil {
				return fmt.Errorf("unable to get prefect client: %w", err)
			}

			// NOTE: the accountID is inherited by the one set in the test environment
			teamsClient, err := client.Teams(uuid.Nil)
			if err != nil {
				return fmt.Errorf("unable to get teams client: %w", err)
			}

			teams, err := teamsClient.List(context.Background(), []string{})
			if err != nil {
				return fmt.Errorf("unable to list teams: %w", err)
			}

			if len(teams) == 0 {
				return fmt.Errorf("no teams found for this account")
			}

			for _, team := range teams {
				if strings.HasPrefix(team.Name, testutils.TestAccPrefix) {
					log.Printf("found acceptance testing team %s, deleting...\n", team.Name)

					err := teamsClient.Delete(context.Background(), team.ID.String())
					if err != nil {
						log.Printf("unable to delete team %s during sweep: %s\n", team.Name, err)
					}
				} else {
					log.Printf("team %s does not match acceptance testing prefix, skipping...\n", team.Name)
				}
			}

			return nil
		},
	})
}
