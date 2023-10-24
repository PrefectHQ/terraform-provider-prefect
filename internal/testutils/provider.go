package testutils

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider"
)

const (
	// ProviderConfig can be imported into each acceptance test case
	// to initialize a provider configuration to be used in each assertion.
	// dynamically inject configurations once provider.go is fixed
	// https://github.com/PrefectHQ/terraform-provider-prefect/issues/67
	ProviderConfig = `
provider "prefect" {}
`
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"prefect": providerserver.NewProtocol6WithError(provider.New()),
}
