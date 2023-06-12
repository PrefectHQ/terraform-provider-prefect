package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
)

// Provider implements the Prefect Terraform provider.
type Provider struct {
	client *client.Client
}

// ProviderModel maps provider schema data to a Go type.
type ProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	APIKey      types.String `tfsdk:"api_key"`
	AccountID   types.String `tfsdk:"account_id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}
