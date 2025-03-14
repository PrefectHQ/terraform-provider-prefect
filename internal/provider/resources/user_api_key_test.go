package resources_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccUserAPIKeyCreate(userID, name string) string {
	return fmt.Sprintf(`
resource "prefect_user_api_key" "test" {
  user_id = "%s"
  name = "%s"
}
`, userID, name)
}

func fixtureAccUserAPIKeyRecreate(userID, name string, expiration time.Time) string {
	return fmt.Sprintf(`
resource "prefect_user_api_key" "test" {
  user_id = "%s"
  name = "%s"
	expiration = "%s"
}
`, userID, name, expiration.Format(time.RFC3339))
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_user_api_key(t *testing.T) {
	resourceName := "prefect_user_api_key.test"
	userID := os.Getenv("ACC_TEST_USER_RESOURCE_ID")
	name := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	expiration := time.Now().Add(time.Hour * 24)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc: SkipIfUserResource,
				Config:   fixtureAccUserAPIKeyCreate(userID, name),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "user_id", userID),
					testutils.ExpectKnownValue(resourceName, "name", name),
				},
			},
			{
				SkipFunc: SkipIfUserResource,
				Config:   fixtureAccUserAPIKeyRecreate(userID, name, expiration),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "user_id", userID),
					testutils.ExpectKnownValue(resourceName, "name", name),
					testutils.ExpectKnownValue(resourceName, "expiration", expiration.Format(time.RFC3339)),
				},
			},
		},
	})
}
