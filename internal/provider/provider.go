package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/datasources"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

var _ = provider.Provider(&PrefectProvider{})

// New returns a new Prefect Provider instance.
//
//nolint:ireturn // required by Terraform API
func New() provider.Provider {
	return &PrefectProvider{}
}

// Metadata returns the provider type name.
func (p *PrefectProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "prefect"
}

// Schema defines the provider-level schema for configuration data.
func (p *PrefectProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Default account ID to act on (Prefect Cloud only)",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Default workspace ID to act on (Prefect Cloud only)",
				Optional:    true,
			},
		},
	}
}

// Configure configures the provider's internal client.
func (p *PrefectProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	config := &PrefectProviderModel{}

	// Populate the model from provider configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure that all configuration values passed in to provider are known
	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#unknown-values
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
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, set the PREFECT_API_KEY environment variable, or remove the value.",
		)
	}

	if config.AccountID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Unknown Prefect Account ID",
			"The Prefect Account ID is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or remove the value.",
		)
	}

	if config.WorkspaceID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("workspace_id"),
			"Unknown Prefect Workspace ID",
			"The Prefect Workspace ID is not known at configuration time. "+
				"Potential resolutions: target apply the source of the value first, set the value statically in the configuration, or remove the value.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract endpoint from configuration or environment variable.
	// If the endpoint is not set, or the value is not a valid URL, emit an error.
	var endpoint string
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	} else if apiURLEnvVar, ok := os.LookupEnv("PREFECT_API_URL"); ok {
		endpoint = apiURLEnvVar
	}
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Prefect API Endpoint",
			"The Prefect API Endpoint is set to an empty value. "+
				"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a non-empty value, or remove the value.",
		)
	}
	// Here, we'll ensure that the /api suffix is present on the endpoint
	if !strings.HasSuffix(endpoint, "/api") {
		endpoint = fmt.Sprintf("%s/api", endpoint)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Invalid Prefect API Endpoint",
			fmt.Sprintf("The Prefect API Endpoint %q is not a valid URL: %s", endpoint, err),
		)
	}
	isPrefectCloudEndpoint := endpointURL.Host == "api.prefect.cloud" || endpointURL.Host == "api.prefect.dev" || endpointURL.Host == "api.stg.prefect.dev"

	// Extract the API Key from configuration or environment variable.
	var apiKey string
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	} else if apiKeyEnvVar, ok := os.LookupEnv("PREFECT_API_KEY"); ok {
		apiKey = apiKeyEnvVar
	}

	// Extract the Account ID from configuration or environment variable.
	// If the ID is set to an invalid UUID, emit an error.
	var accountID uuid.UUID
	if !config.AccountID.IsNull() {
		accountID = config.AccountID.ValueUUID()
	} else if accountIDEnvVar, ok := os.LookupEnv("PREFECT_CLOUD_ACCOUNT_ID"); ok {
		accountID, err = uuid.Parse(accountIDEnvVar)
		if err != nil {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("account_id"),
				"Invalid Prefect Account ID defined in PREFECT_CLOUD_ACCOUNT_ID ",
				fmt.Sprintf("The PREFECT_CLOUD_ACCOUNT_ID value %q is not a valid UUID: %s", accountIDEnvVar, err),
			)
		}
	}

	// If the endpoint is pointed to Prefect Cloud, we will ensure
	// that a valid API Key is passed.
	// Additionally, we will warn if an Account ID is missing,
	// as it's likely that this is a user misconfiguration.
	if isPrefectCloudEndpoint {
		if apiKey == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("api_key"),
				"Missing Prefect API Key",
				"The Prefect API Endpoint is configured to Prefect Cloud, however, the Prefect API Key is empty. "+
					"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a Prefect server installation, set the PREFECT_API_KEY environment variable, or configure the api_key attribute.",
			)
		}

		if accountID == uuid.Nil {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("account_id"),
				"Missing Prefect Account ID",
				"The Prefect API Endpoint is configured to Prefect Cloud, however, the Prefect Account ID is empty. "+
					"Potential resolutions: set the PREFECT_CLOUD_ACCOUNT_ID environment variable, or configure the account_id attribute.",
			)
		}
	}

	// If the endpoint is pointed to a self-hosted Prefect Server installation,
	// we will warn the practitioner if an API Key is set, as it's possible that
	// this is a user misconfiguration.
	if !isPrefectCloudEndpoint {
		if apiKey != "" {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("api_key"),
				"Prefect API Key ",
				"The Prefect API Key is set, however, the Endpoint is set to a Prefect server installation. "+
					"Potential resolutions: set the endpoint attribute or PREFECT_API_URL environment variable to a Prefect Cloud endpoint, unset the PREFECT_API_KEY environment variable, or remove the api_key attribute.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	prefectClient, err := client.New(
		client.WithEndpoint(endpoint),
		client.WithAPIKey(apiKey),
		client.WithDefaults(accountID, config.WorkspaceID.ValueUUID()),
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
func (p *PrefectProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewAccountDataSource,
		datasources.NewServiceAccountDataSource,
		datasources.NewVariableDataSource,
		datasources.NewWorkPoolDataSource,
		datasources.NewWorkPoolsDataSource,
		datasources.NewWorkspaceDataSource,
		datasources.NewWorkspaceRoleDataSource,
		datasources.NewAccountRoleDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *PrefectProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewAccountResource,
		resources.NewServiceAccountResource,
		resources.NewVariableResource,
		resources.NewWorkPoolResource,
		resources.NewWorkspaceAccessResource,
		resources.NewWorkspaceResource,
		resources.NewWorkspaceRoleResource,
	}
}
