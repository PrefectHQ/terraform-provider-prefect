package resources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func TestPointerTimeEqualityHelper(t *testing.T) {
	t.Parallel()
	now := time.Now()
	pointInTime1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	pointInTime2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	pointInTime3 := time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		t1, t2 *time.Time
		want   bool
	}{
		{nil, nil, true},
		{nil, &now, false},
		{&now, nil, false},
		{&pointInTime1, &pointInTime2, true},
		{&pointInTime1, &pointInTime3, false},
	}

	for _, c := range cases {
		got := resources.ArePointerTimesEqual(c.t1, c.t2)
		if got != c.want {
			t.Fatalf("%v == %v should be %v, but got %v", c.t1, c.t2, c.want, got)
		}
	}
}

func fixtureAccServiceAccountResource(name string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
}`, name)
}

func fixtureAccServiceAccountResourceUpdatedKey(name string, expiration time.Time) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
	api_key_expiration = "%s"
}`, name, expiration.Format(time.RFC3339))
}

func fixtureAccServiceAccountResourceAccountRole(name string) string {
	return fmt.Sprintf(`
data "prefect_account_role" "admin" {
	name = "Admin"
}
resource "prefect_service_account" "bot2" {
	name = "%s"
	account_role_name = "Admin"
}`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_service_account(t *testing.T) {
	resourceName := "prefect_service_account.bot"
	resourceName2 := "prefect_service_account.bot2"
	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	randomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	var apiKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the service account resource
				Config: fixtureAccServiceAccountResource(randomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
				),
			},
			{
				// Ensure non-expiration time change does not trigger a key rotation
				Config: fixtureAccServiceAccountResource(randomName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyUnchanged(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "name", randomName2),
				),
			},
			{
				// Ensure that expiration time change does trigger a key rotation
				Config: fixtureAccServiceAccountResourceUpdatedKey(randomName, time.Now().AddDate(0, 0, 1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountAPIKeyRotated(resourceName, &apiKey),
					resource.TestCheckResourceAttr(resourceName, "name", randomName),
				),
			},
			{
				Config: fixtureAccServiceAccountResourceAccountRole(randomName2),
				Check: resource.ComposeTestCheckFunc(
					// @TODO: This is a superficial test, until we can pull in the provider client
					// and actually test the API call to Prefect Cloud
					resource.TestCheckResourceAttrPair(resourceName2, "account_role_name", "data.prefect_account_role.admin", "name"),
				),
			},
		},
	})
}

func testAccCheckServiceAccountResourceExists(n string, passedKey *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		key := rs.Primary.Attributes["api_key"]
		*passedKey = key

		return nil
	}
}

func testAccCheckServiceAccountAPIKeyUnchanged(n string, passedKey *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		key := rs.Primary.Attributes["api_key"]

		if *passedKey != key {
			return fmt.Errorf("key was incorrectly rotated, since the old key=%s is different from new key=%s", *passedKey, key)
		}

		return nil
	}
}

func testAccCheckServiceAccountAPIKeyRotated(n string, passedKey *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		key := rs.Primary.Attributes["api_key"]

		if *passedKey == key {
			return fmt.Errorf("key rotation did not occur correctly, as the old key=%s is the same as the new key=%s", *passedKey, key)
		}

		return nil
	}
}
