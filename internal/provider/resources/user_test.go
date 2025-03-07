package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

const (
	marvinUserID = "48750018-b4c4-4484-8fbf-b61baf3926b5"
)

func fixtureAccUserResource() string {
	return fmt.Sprintf(`
resource "prefect_user" "marvin" {
  id = "%s"

  # This is here as a safeguard to prevent the user from being
  # destroyed when the test is finished. This is primarily handled by the
  # 'fixtureAccUserResourceUnmanage' variable.
  lifecycle {
    prevent_destroy = true
  }
}
	`, marvinUserID)
}

// This is a helper variable to unmanage the user resource.
// Setting `Destroy: false` in the test steps apparently is not enough
// to prevent the resource from being destroyed.
var fixtureAccUserResourceUnmanage = `
removed {
  from = prefect_user.marvin
  lifecycle {
    destroy = false
  }
}
`

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_user(t *testing.T) {
	resourceName := "prefect_user.marvin"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Start by importing the user resource, since one cannot
				// be created via Terraform.
				Config:             fixtureAccUserResource(),
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      marvinUserID,
				ImportStatePersist: true, // persist the state for subsequent test steps
			},
			{
				// Next, verify that importing doesn't change the values.
				Config:            fixtureAccUserResource(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     marvinUserID,
				ImportStateVerify: true,
			},
			{
				// Finally, unmanage the user resource, so that
				// it is not destroyed when the test is finished.
				Config: fixtureAccUserResourceUnmanage,
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
