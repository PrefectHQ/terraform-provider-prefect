package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccAccountMember() string {
	return `
resource "prefect_account" "test" {
  name = "github-ci-tests"
  handle = "github-ci-tests"
}

resource "prefect_account_member" "test" {
	email = "marvin@prefect.io"
	account_id = "bb19c492-73c2-4ecd-9cd7-d82c4aac08e6"
}
`
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_account_member(t *testing.T) {
	resourceName := "prefect_account_member.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Start by importing the account resource, since one cannot
				// be created via Terraform.
				Config:            `resource "prefect_account" "test" {}`,
				ResourceName:      "prefect_account.test",
				ImportState:       true,
				ImportStateId:     os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"),
				ImportStateVerify: false,
			},
			{
				// Next, import the account member resource, which also cannot
				// be created via Terraform.
				Config:                  fixtureAccAccountMember(),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateId:           "bb19c492-73c2-4ecd-9cd7-d82c4aac08e6,email/marvin@prefect.io",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"id"},
			},
			// {
			// 	// Next, confirm the values are properly set.
			// 	Config: fixtureAccAccountMember(),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckAccountMemberExists(resourceName, &accountMember),
			// 	),
			// },
		},
	})
}

// testAccCheckAccountMemberExists is a Custom Check Function that
// verifies that the API object was created correctly.
func testAccCheckAccountMemberExists(accountMemberResourceName string, accountMember *api.AccountMembership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get the account resource we just created from the state
		accountID, err := testutils.GetResourceIDFromState(s, "prefect_account.test")
		if err != nil {
			return fmt.Errorf("error fetching account ID: %w", err)
		}

		// Initialize the client with the associated accountID
		// NOTE: the accountID is inherited by the one set in the test environment
		c, _ := testutils.NewTestClient()
		accountMembershipsClient, _ := c.AccountMemberships(accountID)

		fetchedAccountMember, err := accountMembershipsClient.List(context.Background(), []string{"marvin@prefect.io"})
		if err != nil {
			return fmt.Errorf("error fetching account member: %w", err)
		}

		// Assign the fetched account member to the passed pointer
		// so we can use it in the next test assertion
		*accountMember = *fetchedAccountMember[0]

		return nil
	}
}
