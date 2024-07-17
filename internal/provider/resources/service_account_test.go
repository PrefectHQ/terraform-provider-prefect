package resources_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
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

func fixtureAccServiceAccountResourceUpdateKeyExpiration(name string, expiration time.Time) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
	api_key_expiration = "%s"
}`, name, expiration.Format(time.RFC3339))
}

func fixtureAccServiceAccountResourceUpdateAccountRoleName(name string, roleName string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
	account_role_name = "%s"
}`, name, roleName)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_service_account(t *testing.T) {
	botResourceName := "prefect_service_account.bot"

	botRandomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	botRandomName2 := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	expiration := time.Now().AddDate(0, 0, 1)

	var apiKey string
	var bot api.ServiceAccount

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Check creation + existence of the service account resource
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					textAccCheckServiceAccountAPIKeyStored(botResourceName, &apiKey),
					resource.TestCheckResourceAttr(botResourceName, "name", botRandomName),
				),
			},
			{
				// Ensure non-expiration time change DOESN'T trigger a key rotation
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyUnchanged(botResourceName, &apiKey),
					resource.TestCheckResourceAttr(botResourceName, "name", botRandomName),
				),
			},
			{
				// Ensure that expiration time change DOES trigger a key rotation
				Config: fixtureAccServiceAccountResourceUpdateKeyExpiration(botRandomName, expiration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyRotated(botResourceName, &apiKey),
					resource.TestCheckResourceAttr(botResourceName, "name", botRandomName),
				),
			},
			{
				// Ensure updates of the account role
				Config: fixtureAccServiceAccountResourceUpdateAccountRoleName(botRandomName, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Admin"}),
					resource.TestCheckResourceAttr(botResourceName, "name", botRandomName),
				),
			},
			{
				// Ensure updates of the service account name
				Config: fixtureAccServiceAccountResourceUpdateAccountRoleName(botRandomName2, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName2, AccountRoleName: "Admin"}),
					resource.TestCheckResourceAttr(botResourceName, "name", botRandomName2),
				),
			},
			// Import State checks - import by name
			{
				ImportState:                          true,
				ImportStateId:                        botRandomName2,
				ImportStateIdPrefix:                  "name/",
				ResourceName:                         botResourceName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"api_key", "old_key_expires_in_seconds"},
			},
			// Import State checks - import by ID (default)
			{
				ImportState:             true,
				ResourceName:            botResourceName,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", "old_key_expires_in_seconds"},
			},
		},
	})
}

func testAccCheckServiceAccountResourceExists(serviceAccountResourceName string, bot *api.ServiceAccount) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// find the corresponding state object
		serviceAccountResource, ok := state.RootModule().Resources[serviceAccountResourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", serviceAccountResourceName)
		}

		// Create a new client, and use the default accountID from environment
		c, _ := testutils.NewTestClient()
		serviceAccountClient, _ := c.ServiceAccounts(uuid.Nil)
		fetchedServiceAccount, err := serviceAccountClient.Get(context.Background(), serviceAccountResource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error fetching Service Account: %w", err)
		}
		if fetchedServiceAccount == nil {
			return fmt.Errorf("Service Account not found for ID: %s", serviceAccountResource.Primary.ID)
		}

		*bot = *fetchedServiceAccount

		return nil
	}
}

func testAccCheckServiceAccountValues(fetchedBot *api.ServiceAccount, valuesToCheck *api.ServiceAccount) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if fetchedBot.Name != valuesToCheck.Name {
			return fmt.Errorf("Expected Service Account name %s, got: %s", fetchedBot.Name, valuesToCheck.Name)
		}

		if fetchedBot.AccountRoleName != valuesToCheck.AccountRoleName {
			return fmt.Errorf("Expected Service Account role name %s, got: %s", fetchedBot.AccountRoleName, valuesToCheck.AccountRoleName)
		}

		return nil
	}
}

func textAccCheckServiceAccountAPIKeyStored(resourceName string, passedKey *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// find the corresponding state object
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
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
			return fmt.Errorf("Resource not found in state: %s", n)
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
			return fmt.Errorf("Resource not found in state: %s", n)
		}

		key := rs.Primary.Attributes["api_key"]

		if *passedKey == key {
			return fmt.Errorf("key rotation did not occur correctly, as the old key=%s is the same as the new key=%s", *passedKey, key)
		}

		return nil
	}
}
