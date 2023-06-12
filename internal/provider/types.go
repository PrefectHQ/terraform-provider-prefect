package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
)

// PrefectProvider implements the Prefect Terraform provider.
type PrefectProvider struct {
	client *client.Client
}

// PrefectProviderModel maps provider schema data to a Go type.
type PrefectProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	APIKey      types.String `tfsdk:"api_key"`
	AccountID   types.String `tfsdk:"account_id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}
