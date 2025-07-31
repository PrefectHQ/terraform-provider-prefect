package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_account(t *testing.T) {
	// Accounts are not compatible OSS.
	testutils.SkipTestsIfOSS(t)

	resourceName := "prefect_account.test"

	checkFunc := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 instance state, got %d", len(s))
		}

		account := s[0]

		tests := []struct {
			attribute string
			expected  string
		}{
			{
				attribute: "name",
				expected:  "github-ci-tests",
			},
			{
				attribute: "handle",
				expected:  "github-ci-tests",
			},
			{
				// This value was provided manually in the UI to support this test.
				attribute: "link",
				expected:  "https://github.com/PrefectHQ/terraform-provider-prefect",
			},
			{
				// Billing email is not available in the staging account because Stripe
				// is not configured.
				attribute: "billing_email",
				expected:  "",
			},
		}

		for _, test := range tests {
			if account.Attributes[test.attribute] != test.expected {
				return fmt.Errorf("expected name to be %s, got %s", test.expected, account.Attributes[test.attribute])
			}
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			// Import State checks - import by ID (from environment)
			// NOTE: the prefect_account resource is a little special in that
			// we cannot create an account via API, meaning the TF lifecycle
			// will be challenging to test. Instead, we'll ensure that the
			// resource can be found and properly imported. Note that
			// ImportStateVerify is set to false, as the resource can't be
			// saved to state after a Create. We make up for this by providing
			// a custom ImportStateCheck function to confirm that the retrieved
			// attributes match expectations.
			{
				Config:            `resource "prefect_account" "test" {}`,
				ImportStateId:     os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"),
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateVerify: false,
				ImportStateCheck:  checkFunc,
			},
		},
	})
}
