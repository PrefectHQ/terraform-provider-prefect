package testutils

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	prefectProvider "github.com/prefecthq/terraform-provider-prefect/internal/provider"
)

// TestAccPrefix is the prefix set for all resources created via acceptance testing,
// so that we can easily identify and clean them up in case of flakiness/failures.
const TestAccPrefix = "terraform_acc_"

// TestAccProvider defines the actual Provider, which is used during acceptance testing.
// This is the same Provider that is used by the CLI, and is used by
// custom test functions, primarily to access the underlying HTTP client.
var TestAccProvider provider.Provider = prefectProvider.New()

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"prefect": providerserver.NewProtocol6WithError(TestAccProvider),
}

// AccTestPreCheck is a utility hook, which every test suite will call
// in order to verify if the necessary provider configurations are passed
// through the environment variables.
// https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/testcase#precheck
func AccTestPreCheck(t *testing.T) {
	t.Helper()
	neededVars := []string{"PREFECT_API_URL", "PREFECT_API_KEY", "PREFECT_CLOUD_ACCOUNT_ID"}
	for _, key := range neededVars {
		if v := os.Getenv(key); v == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}
}
