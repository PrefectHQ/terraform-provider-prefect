package resources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

var fixtureAccAccountMember = `
resource "prefect_account" "test" {
  name =   "github-ci-tests"
  handle = "github-ci-tests"
}

resource "prefect_account_member" "test" {
  # This is here as a safeguard to prevent the account member from being
  # destroyed when the test is finished. This is primarily handled by the
  # 'fixtureAccAccountMemberUnmanage' variable.
  lifecycle {
    prevent_destroy = true
  }
}
`

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
	// Account member is not supported in OSS.
	testutils.SkipTestsIfOSS(t)

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
				Config:            fixtureAccAccountMember,
				ResourceName:      "prefect_account.test",
				ImportState:       true,
				ImportStateId:     accountID,
				ImportStateVerify: false,
			},
			{
				// Next, import the account member resource, which also cannot
				// be created via Terraform.
				Config:             fixtureAccAccountMember,
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      resourceEmail,
				ImportStatePersist: true, // persist the state for subsequent test steps
			},
			{
				// Next, verify that importing doesn't change the values.
				Config:                               fixtureAccAccountMember,
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
			// We don't test a resource update here because the account member
			// resource is currently hard-coded to refer to a specific account
			// member.
			//
			// This is because we cannot currently create an account member via
			// the API (and therefore via Terraform).
			//
			// Because of this, two tests running simultaneously would try to
			// modify the same account member resource, which could lead to
			// flaky tests.
			//
			// In the meantime, the resource update has been tested manually
			// by modifying the account role ID in a local Terraform configuration
			// and confirming that the role was updated in the UI as expected.
		},
	})
}
