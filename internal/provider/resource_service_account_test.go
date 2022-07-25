package provider_test

import (
	"context"
	"fmt"
	"regexp"
	"terraform-provider-prefect/api"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("service_account", &resource.Sweeper{
		Name: "service_account",
		F:    testSweepServiceAccounts,
	})
}

func testSweepServiceAccounts(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	users, err := api.GetTenantUserIdsWithSimilarUsername(client.GQLClient, context.Background(), "testacc%")
	if err != nil {
		return fmt.Errorf("Could not read service accounts: %v", err)
	}

	for _, user := range users {
		result, err := api.DeleteServiceAccount(client.GQLClient, context.Background(), (api.UUID)(*user.Id))
		if err != nil {
			return fmt.Errorf("Could not delete service account: %v (service account ID = %s)", err, *user.Id)
		}
		if !*result.Success {
			return fmt.Errorf("Could not delete service account: %v (service account ID = %s)", *result.Error, *user.Id)
		}
	}

	return nil
}

func TestAccServiceAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccServiceAccountResourceConfig("testacc-service-account", "USER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_service_account.test", "name", "testacc-service-account"),
					resource.TestCheckResourceAttr("prefect_service_account.test", "role", "USER"),
				),
			},
			// Update role and refresh
			{
				Config: testAccServiceAccountResourceConfig("testacc-service-account", "READ_ONLY_USER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_service_account.test", "name", "testacc-service-account"),
					resource.TestCheckResourceAttr("prefect_service_account.test", "role", "READ_ONLY_USER"),
				),
			},
			// Update name should fail
			{
				Config:      testAccServiceAccountResourceConfig("new-name", "READ_ONLY_USER"),
				ExpectError: regexp.MustCompile(`Prefect does not allow admins to change user names`),
			},
			// Import and refresh
			{
				ResourceName:      "prefect_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccServiceAccountResourceConfig(name string, role string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "test" {
	name = %[1]q
	role = %[2]q
}
`, name, role)
}

// nolint:funlen
func TestAccServiceAccountResourceApiKeys(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// API key names must be unique
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
						  name = "key1"
						},
						{
						  name = "key1"
						},
					  ]
				}`,
				ExpectError: regexp.MustCompile(`api key name is not unique`),
			},
			// Timezone must be +00:00
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
							name = "key1"
							expiration = "2015-10-21T16:29:00+13:00"
						},
						]
				}`,
				ExpectError: regexp.MustCompile(`must be in the \+00:00 timezone`),
			},
			// Create
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
							name       = "key1"
							expiration = "2015-10-21T16:29:00+00:00"
						},
					]
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.0.name", "key1"),
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.0.expiration", "2015-10-21T16:29:00+00:00"),
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.0.id"),
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.0.key"),
				),
			},
			// Cannot update expiration
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
							name       = "key1"
							expiration = "2020-10-21T16:29:00+00:00"
						},
					]
				}`,
				ExpectError: regexp.MustCompile(`api key expiration cannot be changed after creation`),
			},
			// Update: Add new key
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
							name       = "key1"
							expiration = "2015-10-21T16:29:00+00:00"
						},
						{
							name       = "key2"
							expiration = "2030-10-21T16:29:00+00:00"
						}
					]
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.1.name", "key2"),
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.1.expiration", "2030-10-21T16:29:00+00:00"),
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.1.id"),
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.1.key"),
				),
			},
			// Update: Replace/rotate key
			{
				Config: `
				resource "prefect_service_account" "test" {
					name = "testacc-service-account"
					role = "USER"
					api_keys = [
						{
							name       = "key1"
							expiration = "2015-10-21T16:29:00+00:00"
						},
						{
							name       = "key3"
							expiration = "2040-10-21T16:29:00+00:00"
						}
					]
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.1.name", "key3"),
					resource.TestCheckResourceAttr("prefect_service_account.test", "api_keys.1.expiration", "2040-10-21T16:29:00+00:00"),
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.1.id"),
					// this will be a different key
					resource.TestCheckResourceAttrSet("prefect_service_account.test", "api_keys.1.key"),
				),
			},
			// Import and refresh
			{
				ResourceName:      "prefect_service_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				// ignore keys because they're only ever returned once on creation and can never be fetched later ie: during an import
				ImportStateVerifyIgnore: []string{"api_keys.0.key", "api_keys.1.key"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
