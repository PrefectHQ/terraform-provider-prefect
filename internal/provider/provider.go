package provider

import (
	"context"
	"fmt"
	"os"
	"terraform-provider-prefect/internal/prefect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

type provider struct {
	configured bool
	client     *prefect.Client

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GetSchema - defines the provider's attributes
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{

		MarkdownDescription: `
Welcome to the [Prefect](https://prefect.io) provider!

Use the navigation to the left to read about the available resources.
`,

		Attributes: map[string]tfsdk.Attribute{
			"api_server": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				MarkdownDescription: fmt.Sprintf("Prefect API url. If unspecified the env var `PREFECT__CLOUD__API` will be used. If neither are set will default to [%s](%s)",
					prefect.DefaultAPIServer,
					prefect.DefaultAPIServer),
			},
			"api_key": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Prefect API key. The api key used determines the Prefect tenant. If unspecified the env var `PREFECT__CLOUD__API_KEY` will be used.",
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	APIServer types.String `tfsdk:"api_server"`
	APIKey    types.String `tfsdk:"api_key"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiServer string
	if config.APIServer.Null {
		apiServer = os.Getenv("PREFECT__CLOUD__API")
	} else {
		apiServer = config.APIServer.Value
	}

	if apiServer == "" {
		apiServer = prefect.DefaultAPIServer
	}

	var apiKey string
	if config.APIKey.Null {
		apiKey = os.Getenv("PREFECT__CLOUD__API_KEY")
	} else {
		apiKey = config.APIKey.Value
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Unable to find api key",
			"api_key or the env var PREFECT__CLOUD__API_KEY must be specified",
		)
		return
	}

	// Create a new Prefect client and set it to the provider client
	c, err := prefect.NewClient(ctx, &apiKey, &apiServer)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create prefect client",
			err.Error(),
		)
		return
	}

	p.client = c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"prefect_project":         projectResourceType{},
		"prefect_service_account": serviceAccountResourceType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf(`While creating the data source or resource, an unexpected provider type (%T) was received.
 This is always a bug in the provider code and should be reported to the provider developers.`, p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			`While creating the data source or resource, an unexpected empty provider instance was received.
 This is always a bug in the provider code and should be reported to the provider developers.`,
		)
		return provider{}, diags
	}

	return *p, diags
}
