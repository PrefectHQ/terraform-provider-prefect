package testutils

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	prefectProvider "github.com/prefecthq/terraform-provider-prefect/internal/provider"
)

const (
	// TestAccPrefix is the prefix set for all resources created via acceptance testing,
	// so that we can easily identify and clean them up in case of flakiness/failures.
	TestAccPrefix = "terraformacc"

	// RandomStringLength sets the length of the random string used when creating a new random
	// name for a resource via NewRandomEphemeralWorkspace.
	RandomStringLength = 10

	// WorkspaceResourceName is the name of the workspace resource.
	WorkspaceResourceName = "prefect_workspace.test"
)

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

func TestContextOSS() bool {
	return os.Getenv("TEST_CONTEXT") == "OSS"
}

// AccTestPreCheck is a utility hook, which every test suite will call
// in order to verify if the necessary provider configurations are passed
// through the environment variables.
// https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/testcase#precheck
func AccTestPreCheck(t *testing.T) {
	t.Helper()

	// Exit early if we're testing against Prefect OSS.
	if TestContextOSS() {
		return
	}

	neededVars := []string{"PREFECT_API_URL", "PREFECT_API_KEY", "PREFECT_CLOUD_ACCOUNT_ID"}
	for _, key := range neededVars {
		if v := os.Getenv(key); v == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}
}

// NewTestClient returns a new Prefect API client instance
// to be used in acceptance tests.
// The plugin-framework does not currently expose a way to extract
// the provider-configured client - so instead, we duplicate some
// of the client initiatlization logic that also happens in Provider.Configure().
// https://github.com/hashicorp/terraform-plugin-testing/issues/11
//
//nolint:ireturn // required by Terraform API
func NewTestClient() (api.PrefectClient, error) {
	endpoint := os.Getenv("PREFECT_API_URL")
	apiKey := os.Getenv("PREFECT_API_KEY")

	if !strings.HasSuffix(endpoint, "/api") {
		endpoint = fmt.Sprintf("%s/api", endpoint)
	}

	endpointURL, _ := url.Parse(endpoint)
	endpointHost := fmt.Sprintf("%s://%s", endpointURL.Scheme, endpointURL.Host)

	var prefectClient *client.Client

	opts := []client.Option{
		client.WithEndpoint(endpoint, endpointHost),
		client.WithAPIKey(apiKey),
	}

	if !TestContextOSS() {
		aID := os.Getenv("PREFECT_CLOUD_ACCOUNT_ID")
		accountID, _ := uuid.Parse(aID)
		opts = append(opts, client.WithDefaults(accountID, uuid.Nil))
	}

	prefectClient, _ = client.New(opts...)

	return prefectClient, nil
}

// NewRandomPrefixedString returns a new random prefixed string used for creating ephemeral resources.
func NewRandomPrefixedString() string {
	return TestAccPrefix + acctest.RandStringFromCharSet(RandomStringLength, acctest.CharSetAlphaNum)
}

// Workspace is a struct that represents a workspace for acceptance tests.
type Workspace struct {
	Resource    string
	Name        string
	Description string
}

// NewEphemeralWorkspace returns a new ephemeral workspace for acceptance tests.
func NewEphemeralWorkspace() Workspace {
	workspace := Workspace{}

	randomName := NewRandomPrefixedString()
	workspace.Name = randomName
	workspace.Description = randomName

	workspace.Resource = fmt.Sprintf(`
resource "prefect_workspace" "test" {
	name = "%s"
	handle = "%s"
	description = "%s"
}`, randomName, randomName, randomName)

	return workspace
}
