package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
)

// Provider implements the Prefect Terraform provider.
type Provider struct {
	client *client.Client
}

// providerModel maps provider schema data to a Go type.
type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}
