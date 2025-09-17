package resources_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
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

func fixtureAccServiceAccountResourceKeyKeepers(name string, keeperValue string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "bot" {
	name = "%s"
	api_key_keepers = {
	  foo = "%s"
	}
}`, name, keeperValue)
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
	// Service accounts are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	botResourceName := "prefect_service_account.bot"

	botRandomName := testutils.NewRandomPrefixedString()
	botRandomName2 := testutils.NewRandomPrefixedString()

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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
				},
			},
			{
				// Ensure non-expiration time change DOESN'T trigger a key rotation
				Config: fixtureAccServiceAccountResource(botRandomName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName2, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyUnchanged(botResourceName, &apiKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName2),
				},
			},
			{
				// Ensure that expiration time change DOES trigger a key rotation
				Config: fixtureAccServiceAccountResourceUpdateKeyExpiration(botRandomName, expiration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyRotated(botResourceName, &apiKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
				},
			},
			{
				// Ensure that switching to key keepers DOES trigger a key rotation
				Config: fixtureAccServiceAccountResourceKeyKeepers(botRandomName, "keeper-value-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyRotated(botResourceName, &apiKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
				},
			},
			{
				// Ensure that key keepers change DOES trigger a key rotation
				Config: fixtureAccServiceAccountResourceKeyKeepers(botRandomName, "keeper-value-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyRotated(botResourceName, &apiKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
				},
			},
			{
				// Ensure that a non-key keeper change DOES NOT trigger a key rotation
				Config: fixtureAccServiceAccountResourceKeyKeepers(botRandomName2, "keeper-value-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName2, AccountRoleName: "Member"}),
					testAccCheckServiceAccountAPIKeyUnchanged(botResourceName, &apiKey),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName2),
				},
			},
			{
				// Ensure updates of the account role
				Config: fixtureAccServiceAccountResourceUpdateAccountRoleName(botRandomName, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Admin"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
				},
			},
			{
				// Ensure updates of the service account name
				Config: fixtureAccServiceAccountResourceUpdateAccountRoleName(botRandomName2, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName2, AccountRoleName: "Admin"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName2),
				},
			},
			{
				// Import State checks - import by name
				ImportState:                          true,
				ImportStateId:                        botRandomName2,
				ImportStateIdPrefix:                  "name/",
				ResourceName:                         botResourceName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateVerifyIgnore:              []string{"api_key", "old_key_expires_in_seconds"},
			},
			{
				// Import State checks - import by ID (default)
				ImportState:             true,
				ResourceName:            botResourceName,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", "old_key_expires_in_seconds"},
			},
		},
	})
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_service_account_external_rotation_detection(t *testing.T) {
	// NOTE: This test validates the behavior and state consistency of the external
	// key rotation detection logic. Due to limitations in the terraform-plugin-testing
	// framework (https://github.com/hashicorp/terraform-plugin-testing/issues/69),
	// we cannot directly test warning diagnostics in acceptance tests.
	// This test focuses on ensuring state remains consistent when the detection
	// logic executes during Read operations.
	testutils.SkipTestsIfOSS(t)

	botResourceName := "prefect_service_account.bot"
	botRandomName := testutils.NewRandomPrefixedString()

	var bot api.ServiceAccount

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountValues(&bot, &api.ServiceAccount{Name: botRandomName, AccountRoleName: "Member"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
					testutils.ExpectKnownValueNotNull(botResourceName, "api_key_id"),
					testutils.ExpectKnownValueNotNull(botResourceName, "api_key_name"),
					testutils.ExpectKnownValueNotNull(botResourceName, "api_key_created"),
				},
			},
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountStateConsistency(botResourceName, &bot),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(botResourceName, "name", botRandomName),
					testutils.ExpectKnownValueNotNull(botResourceName, "api_key_id"),
				},
			},
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountAPIKeyMetadataConsistent(botResourceName, &bot),
				),
			},
		},
	})
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_service_account_read_operation_stability(t *testing.T) {
	testutils.SkipTestsIfOSS(t)

	botResourceName := "prefect_service_account.bot"
	botRandomName := testutils.NewRandomPrefixedString()

	var bot api.ServiceAccount
	var apiKey string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					textAccCheckServiceAccountAPIKeyStored(botResourceName, &apiKey),
				),
			},
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountAPIKeyUnchanged(botResourceName, &apiKey),
					testAccCheckServiceAccountReadStability(botResourceName, &bot),
				),
			},
			{
				Config: fixtureAccServiceAccountResource(botRandomName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceAccountResourceExists(botResourceName, &bot),
					testAccCheckServiceAccountAPIKeyUnchanged(botResourceName, &apiKey),
				),
			},
		},
	})
}

func testAccCheckServiceAccountResourceExists(serviceAccountResourceName string, bot *api.ServiceAccount) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// find the corresponding state object
		serviceAccountResourceID, err := testutils.GetResourceIDFromState(state, serviceAccountResourceName)
		if err != nil {
			return fmt.Errorf("error fetching service account ID: %w", err)
		}

		// Create a new client, and use the default accountID from environment
		c, _ := testutils.NewTestClient()
		serviceAccountClient, _ := c.ServiceAccounts(uuid.Nil)
		fetchedServiceAccount, err := serviceAccountClient.Get(context.Background(), serviceAccountResourceID.String())
		if err != nil {
			return fmt.Errorf("Error fetching Service Account: %w", err)
		}
		if fetchedServiceAccount == nil {
			return fmt.Errorf("Service Account not found for ID: %s", serviceAccountResourceID)
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

// testAccCheckServiceAccountAPIKeyUnchanged is a helper function that checks if the API key was unchanged.
// Upon success, it will ensure that the passeKey is updated to the state key (which is a no-op).
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
		*passedKey = key

		return nil
	}
}

// testAccCheckServiceAccountAPIKeyRotated is a helper function that checks if the API key was rotated correctly.
// Upon success, it will ensure that the passeKey is updated to the state key (which is new).
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
		*passedKey = key

		return nil
	}
}

// testAccCheckServiceAccountStateConsistency verifies that the Terraform state
// is consistent with the API response for key metadata fields.
func testAccCheckServiceAccountStateConsistency(resourceName string, fetchedBot *api.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		stateAPIKeyID := rs.Primary.Attributes["api_key_id"]
		stateAPIKeyName := rs.Primary.Attributes["api_key_name"]

		if stateAPIKeyID != fetchedBot.APIKey.ID {
			return fmt.Errorf("State api_key_id (%s) does not match API response (%s)", stateAPIKeyID, fetchedBot.APIKey.ID)
		}

		if stateAPIKeyName != fetchedBot.APIKey.Name {
			return fmt.Errorf("State api_key_name (%s) does not match API response (%s)", stateAPIKeyName, fetchedBot.APIKey.Name)
		}

		return nil
	}
}

// testAccCheckServiceAccountAPIKeyMetadataConsistent verifies that the API key metadata
// in state matches what was fetched from the API, ensuring the external rotation
// detection logic properly updates state.
func testAccCheckServiceAccountAPIKeyMetadataConsistent(resourceName string, fetchedBot *api.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		stateAPIKeyID := rs.Primary.Attributes["api_key_id"]
		stateAPIKeyCreated := rs.Primary.Attributes["api_key_created"]

		if stateAPIKeyID != fetchedBot.APIKey.ID {
			return fmt.Errorf("API Key ID mismatch - state: %s, API: %s", stateAPIKeyID, fetchedBot.APIKey.ID)
		}

		if fetchedBot.APIKey.Created != nil {
			expectedCreated := fetchedBot.APIKey.Created.Format(time.RFC3339)
			if stateAPIKeyCreated != expectedCreated {
				return fmt.Errorf("API Key creation time mismatch - state: %s, API: %s", stateAPIKeyCreated, expectedCreated)
			}
		}

		return nil
	}
}

// testAccCheckServiceAccountReadStability ensures that repeated read operations
// maintain consistent state and don't introduce false positive external rotation warnings.
func testAccCheckServiceAccountReadStability(resourceName string, fetchedBot *api.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found in state: %s", resourceName)
		}

		stateAPIKeyID := rs.Primary.Attributes["api_key_id"]
		if stateAPIKeyID == "" {
			return fmt.Errorf("API Key ID should not be empty in state")
		}

		if stateAPIKeyID != fetchedBot.APIKey.ID {
			return fmt.Errorf("Read operation instability detected - state key ID changed unexpectedly")
		}

		stateUpdated := rs.Primary.Attributes["updated"]
		if fetchedBot.Updated != nil {
			expectedUpdated := fetchedBot.Updated.Format(time.RFC3339)
			if stateUpdated != expectedUpdated {
				return fmt.Errorf("Service account updated timestamp mismatch - state: %s, API: %s", stateUpdated, expectedUpdated)
			}
		}

		return nil
	}
}
