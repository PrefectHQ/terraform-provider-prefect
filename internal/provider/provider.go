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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/datasources"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
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
				Description: "Prefect API URL. Can also be set via the `PREFECT_API_URL` environment variable. Defaults to `https://api.prefect.cloud`",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "Prefect Cloud API Key. Can also be set via the `PREFECT_API_KEY` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Default Prefect Cloud Account ID. Can also be set via the `PREFECT_CLOUD_ACCOUNT_ID` environment variable.",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Default Prefect Cloud Workspace ID.",
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
		endpoint = "https://api.prefect.cloud"
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
	isPrefectCloudEndpoint := helpers.IsCloudEndpoint(endpointURL.Host)

	// Extracts the host (without the /api suffix),
	// so we can store it on the Client object in addition to the endpoint.
	// For non-Cloud endpoints, it will likely be the same as .endpoint.
	// This is useful for certain resources where we need access to the
	// endpoint host to construct custom URLs as a resource attribute.
	endpointHost := fmt.Sprintf("%s://%s", endpointURL.Scheme, endpointURL.Host)

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

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "prefect_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "prefect_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "prefect_api_key")
	ctx = tflog.SetField(ctx, "prefect_account_id", accountID)
	ctx = tflog.SetField(ctx, "prefect_workspace_id", config.WorkspaceID.ValueString())
	tflog.Debug(ctx, "Creating Prefect client")

	prefectClient, err := client.New(
		client.WithEndpoint(endpoint, endpointHost),
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

	tflog.Info(ctx, "Configured Prefect client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *PrefectProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewAccountDataSource,
		datasources.NewAccountMemberDataSource,
		datasources.NewAccountMembersDataSource,
		datasources.NewAccountRoleDataSource,
		datasources.NewAutomationDataSource,
		datasources.NewBlockDataSource,
		datasources.NewDeploymentDataSource,
		datasources.NewServiceAccountDataSource,
		datasources.NewTeamDataSource,
		datasources.NewTeamsDataSource,
		datasources.NewVariableDataSource,
		datasources.NewWebhookDataSource,
		datasources.NewWorkerMetadataDataSource,
		datasources.NewWorkPoolDataSource,
		datasources.NewWorkPoolsDataSource,
		datasources.NewWorkspaceDataSource,
		datasources.NewWorkspaceRoleDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *PrefectProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewAccountResource,
		resources.NewAutomationResource,
		resources.NewBlockAccessResource,
		resources.NewBlockResource,
		resources.NewDeploymentAccessResource,
		resources.NewDeploymentResource,
		resources.NewDeploymentScheduleResource,
		resources.NewFlowResource,
		resources.NewServiceAccountResource,
		resources.NewVariableResource,
		resources.NewWebhookResource,
		resources.NewWorkPoolResource,
		resources.NewWorkspaceAccessResource,
		resources.NewWorkspaceResource,
		resources.NewWorkspaceRoleResource,
	}
}
