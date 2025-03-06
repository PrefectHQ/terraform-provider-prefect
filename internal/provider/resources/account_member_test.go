package resources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

  # This is here as a safeguard to prevent the account member from being
  # destroyed when the test is finished. This is primarily handled by the
  # 'fixtureAccAccountMemberUnmanage' variable.
  lifecycle {
    prevent_destroy = true
  }
}
`
}

// This is a helper variable to unmanage the account member resource.
// Setting `Destroy: false` in the test steps apparently is not enough
// to prevent the resource from being destroyed.
var fixtureAccAccountMemberUnmanage = `
removed {
  from = prefect_account_member.test
  lifecycle {
    destroy = false
  }
}
`

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_account_member(t *testing.T) {
	resourceName := "prefect_account_member.test"
	resourceEmail := "marvin@prefect.io"
	accountID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Start by importing the account resource, since one cannot
				// be created via Terraform.
				Config:            fixtureAccAccountMember(),
				ResourceName:      "prefect_account.test",
				ImportState:       true,
				ImportStateId:     accountID,
				ImportStateVerify: false,
			},
			{
				// Next, import the account member resource, which also cannot
				// be created via Terraform.
				Config:             fixtureAccAccountMember(),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      resourceEmail,
				ImportStatePersist: true, // persist the state for subsequent test steps
			},
			{
				// Next, verify that importing doesn't change the values.
				Config:                               fixtureAccAccountMember(),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        resourceEmail,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "email",
			},
			{
				// Finally, unmanage the account member resource, so that
				// it is not destroyed when the test is finished.
				Config: fixtureAccAccountMemberUnmanage,
			},
		},
	})
}
