package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/datasources"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

var _ provider.Provider = &Provider{}

// New returns a new Prefect Provider instance.
//
//nolint:ireturn // required by Terraform API
func New() provider.Provider {
	return &Provider{}
}

// Metadata returns the provider type name.
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "prefect"
}

// Schema defines the provider-level schema for configuration data.
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "URL prefix for Prefect Server or Prefect Cloud",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API Key for authenticating to the server (Prefect Cloud only)",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure configures the provider's internal client.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	config := &providerModel{}

	// Populate the model from provider configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure that any values passed in to provider are known
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown Prefect API Endpoint",
			"The Prefect API Endpoint is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or set the PREFECT_API_URL environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Prefect API Key",
			"The Prefect API Key is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, set the PREFECT_API_URL environment variable, or remove the value.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider values if supplied, otherwise fall back to environment variables
	var endpoint string
	if !config.Endpoint.IsUnknown() && !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	} else if u, ok := os.LookupEnv("PREFECT_API_URL"); ok {
		endpoint = u
	} else {
		endpoint = "http://localhost:4200/api"
	}

	var apiKey string
	if !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	} else if key, ok := os.LookupEnv("PREFECT_API_KEY"); ok {
		apiKey = key
	}

	// Validate values (ensure that they are non-empty)
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Prefect API Endpoint",
			"The Prefect API Endpoint is set to an empty value. "+
				"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a non-empty value, or remove the value. "+fmt.Sprintf("endpoint %q unknown %t", endpoint, config.Endpoint.IsUnknown()),
		)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Invalid Prefect API Endpoint",
			fmt.Sprintf("The Prefect API Endpoint %q is not a valid URL: %s", endpoint, err),
		)
	}

	// If API Key is unset, check that we're running against Prefect Cloud
	if endpointURL.Host == "api.prefect.cloud" || endpointURL.Host == "api.prefect.dev" || endpointURL.Host == "api.stg.prefect.dev" {
		if apiKey == "" {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("api_key"),
				"Missing Prefect API Key",
				"The Prefect API Endpoint is configured to Prefect Cloud, however, the Prefect API Key is empty. "+
					"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a Prefect server installation, set the PREFECT_API_KEY environment variable, or configure the api_key attribute.",
			)
		}
	} else if apiKey != "" {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("api_key"),
			"Non-Empty Prefect API Key",
			"The Prefect API Key is set, however, the Endpoint is set to a Prefect server installation. "+
				"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a Prefect Cloud endpoint, unset the PREFECT_API_KEY environment variable, or remove the api_key attribute.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	prefectClient, err := client.New(
		client.WithEndpoint(endpoint),
		client.WithAPIKey(apiKey),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Prefect API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Prefect API client. This is a bug in the provider, please create an issue against https://github.com/PrefectHQ/terraform-provider-prefect unless it has already been reported. "+
				"Error returned by the client: %s", err),
		)

		return
	}
	p.client = prefectClient

	// Pass client to DataSource and Resource type Configure methods
	resp.DataSourceData = prefectClient
	resp.ResourceData = prefectClient
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewVariableDataSource,
		datasources.NewWorkPoolDataSource,
		datasources.NewWorkPoolsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewVariableResource,
		resources.NewWorkPoolResource,
	}
}
