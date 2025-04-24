package resources_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_account(t *testing.T) {
	// Accounts are not compatible OSS.
	testutils.SkipTestsIfOSS(t)

	resourceName := "prefect_account.test"

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
			// saved to state after a Create.
			{
				Config:            `resource "prefect_account" "test" {}`,
				ImportStateId:     os.Getenv("PREFECT_CLOUD_ACCOUNT_ID"),
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateVerify: false,
			},
		},
	})
}
