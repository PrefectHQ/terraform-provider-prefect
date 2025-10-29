package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_account_settings(t *testing.T) {
	// Accounts are not compatible OSS.
	testutils.SkipTestsIfOSS(t)

	accountID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")
	resourceName := "prefect_account.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: Import the account - persist state for subsequent steps
			{
				Config: `
resource "prefect_account" "test" {
	name         = "github-ci-tests"
	handle       = "github-ci-tests"
	link         = "https://github.com/PrefectHQ/terraform-provider-prefect"
	domain_names = ["example.com", "foobar.com"]
}
`,
				ImportStateId:      accountID,
				ImportState:        true,
				ImportStatePersist: true, // persist the state for subsequent test steps
				ResourceName:       resourceName,
				ImportStateVerify:  false,
			},
			// Step 2: Enable enforce_webhook_authentication
			{
				Config: `
resource "prefect_account" "test" {
	name         = "github-ci-tests"
	handle       = "github-ci-tests"
	link         = "https://github.com/PrefectHQ/terraform-provider-prefect"
	domain_names = ["example.com", "foobar.com"]

	settings = {
		enforce_webhook_authentication = true
	}
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueBool(resourceName, "settings.enforce_webhook_authentication", true),
				},
			},
			// Step 3: Disable enforce_webhook_authentication
			{
				Config: `
resource "prefect_account" "test" {
	name         = "github-ci-tests"
	handle       = "github-ci-tests"
	link         = "https://github.com/PrefectHQ/terraform-provider-prefect"
	domain_names = ["example.com", "foobar.com"]

	settings = {
		enforce_webhook_authentication = false
	}
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueBool(resourceName, "settings.enforce_webhook_authentication", false),
				},
			},
			// Step 4: Re-enable and verify it works with all settings together
			{
				Config: `
resource "prefect_account" "test" {
	name         = "github-ci-tests"
	handle       = "github-ci-tests"
	link         = "https://github.com/PrefectHQ/terraform-provider-prefect"
	domain_names = ["example.com", "foobar.com"]

	settings = {
		enforce_webhook_authentication = true
		allow_public_workspaces        = false
		ai_log_summaries               = true
		managed_execution              = true
	}
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValueBool(resourceName, "settings.enforce_webhook_authentication", true),
					testutils.ExpectKnownValueBool(resourceName, "settings.allow_public_workspaces", false),
					testutils.ExpectKnownValueBool(resourceName, "settings.ai_log_summaries", true),
					testutils.ExpectKnownValueBool(resourceName, "settings.managed_execution", true),
				},
			},
			// Note: We cannot include a final step to unmanage the resource using a
			// 'removed' block because Terraform will attempt to destroy it during the
			// apply phase of that step, which fails since accounts cannot be deleted
			// via the API. The post-test cleanup failure is expected and acceptable
			// since the test has already verified the core functionality in steps 1-4.
		},
	})
}
