package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccUserResource(userID string) string {
	return fmt.Sprintf(`
resource "prefect_user" "test" {
  id = "%sinvalid"
}
`, userID)
}

// This is a helper variable to unmanage the user resource
// so the acceptance test framework does not attempt to
// destroy it.
var fixtureAccUserResourceUnmanage = `
removed {
  from = prefect_user.test
}
`

// skipIfUserResource is used to decide whether to skip
// acceptance tests for the User resource.
//
// The default behavior is to skip the test because of the
// caveats noted in the resource documentation.
//
// To override this and manually test, do the following:
//  1. Provide a value for `api_key` from an API key generated as a user,
//     not a service account.
//  2. Provide the user ID as an environment variable: `ACC_TEST_USER_RESOURCE_ID`
func SkipIfUserResource() (bool, error) {
	if os.Getenv("ACC_TEST_USER_RESOURCE") == "yes" {
		return false, nil
	}

	return true, nil
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_user(t *testing.T) {
	resourceName := "prefect_user.test"
	userID := os.Getenv("ACC_TEST_USER_RESOURCE_ID")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Start by importing the user resource, since one cannot
				// be created via Terraform.
				SkipFunc:           SkipIfUserResource,
				Config:             fixtureAccUserResource(userID),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      userID,
				ImportStatePersist: true, // persist the state for subsequent test steps
			},
			{
				// Next, verify that importing doesn't change the values.
				SkipFunc:          SkipIfUserResource,
				Config:            fixtureAccUserResource(userID),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     userID,
				ImportStateVerify: true,
			},
			{
				// Finally, unmanage the user resource, so that
				// it is not destroyed when the test is finished.
				SkipFunc: SkipIfUserResource,
				Config:   fixtureAccUserResourceUnmanage,
			},
			// We don't test a resource update here because the user
			// resource is currently hard-coded to refer to a specific user.
			//
			// This is because we cannot currently create a user via
			// the API (and therefore via Terraform).
			//
			// Because of this, two tests running simultaneously would try to
			// modify the same user resource, which could lead to
			// flaky tests.
			//
			// In the meantime, the resource update has been tested manually
			// by modifying the various fields in a local Terraform configuration
			// and confirming that the user was updated in the UI as expected.
		},
	})
}
