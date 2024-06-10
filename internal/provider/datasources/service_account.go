package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var _ = datasource.DataSourceWithConfigure(&ServiceAccountDataSource{})

// ServiceAccountDataSource contains state for the data source.
type ServiceAccountDataSource struct {
	client api.PrefectClient
}

// ServiceAccountDataSourceModel defines the Terraform data source model.
type ServiceAccountDataSourceModel struct {
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name            types.String          `tfsdk:"name"`
	ActorID         customtypes.UUIDValue `tfsdk:"actor_id"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	AccountRoleName types.String          `tfsdk:"account_role_name"`

	// SA fields
	APIKeyID      types.String               `tfsdk:"api_key_id"`
	APIKeyName    types.String               `tfsdk:"api_key_name"`
	APIKeyCreated customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpires customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKey        types.String               `tfsdk:"api_key"`
}

// NewServiceAccountDataSource returns a new ServiceAccountDataSource.
//
//nolint:ireturn // required by Terraform API
func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

// Metadata returns the data source type name.
func (d *ServiceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

// Configure initializes runtime state for the data source.
func (d *ServiceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

var serviceAccountAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Service Account ID (UUID)",
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Timestamp of when the resource was created (RFC3339)",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Timestamp of when the resource was updated (RFC3339)",
	},
	"actor_id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Actor ID (UUID), used for granting access to resources like Blocks and Deployments",
	},
	"account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID), defaults to the account set in the provider",
		Optional:    true,
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
		Description: "Name of the service account",
	},
	"account_role_name": schema.StringAttribute{
		Computed:    true,
		Description: "Account Role name of the service account",
	},
	"api_key_id": schema.StringAttribute{
		Computed:    true,
		Description: "API Key ID associated with the service account. NOTE: this is always null for reads. If you need the API Key ID, use the `prefect_service_account` resource instead.",
	},
	"api_key_name": schema.StringAttribute{
		Computed:    true,
		Description: "API Key Name associated with the service account",
	},
	"api_key_created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the API Key was created in RFC 3339 format",
	},
	"api_key_expiration": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the API Key expires in RFC 3339 format",
	},
	"api_key": schema.StringAttribute{
		Computed:    true,
		Description: "API Key associated with the service account",
	},
}

// Schema defines the schema for the data source.
func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Service Account, by name or ID.
<br>
Use this data source to obtain service account-level attributes, such as ID.
`,
		Attributes: serviceAccountAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model ServiceAccountDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() && model.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"Either a Service Account ID or Name are required to read a Service Account.",
		)

		return
	}

	client, err := d.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable client",
			fmt.Sprintf("Could not create variable client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	// A Service Account can be read by either ID or Name.
	// If both are set, we prefer the ID
	var serviceAccount *api.ServiceAccount
	if !model.ID.IsNull() {
		serviceAccount, err = client.Get(ctx, model.ID.ValueString())
	} else if !model.Name.IsNull() {
		var serviceAccounts []*api.ServiceAccount
		serviceAccounts, err = client.List(ctx, []string{model.Name.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific service account is not returned
		if err == nil && len(serviceAccounts) != 1 {
			err = fmt.Errorf("a Service Account with the name=%s could not be found", model.Name.ValueString())
		}

		if len(serviceAccounts) == 1 {
			serviceAccount = serviceAccounts[0]
		}
	}

	if serviceAccount == nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not find Service Account with ID=%s and Name=%s", model.ID.ValueString(), model.Name.ValueString()),
		)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(serviceAccount.ID)
	model.Created = customtypes.NewTimestampPointerValue(serviceAccount.Created)
	model.Updated = customtypes.NewTimestampPointerValue(serviceAccount.Updated)

	model.Name = types.StringValue(serviceAccount.Name)
	model.AccountID = customtypes.NewUUIDValue(serviceAccount.AccountID)

	model.AccountRoleName = types.StringValue(serviceAccount.AccountRoleName)
	model.APIKeyID = types.StringValue(serviceAccount.APIKey.ID)
	model.APIKeyName = types.StringValue(serviceAccount.APIKey.Name)
	model.APIKeyCreated = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Created)
	model.APIKeyExpires = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Expiration)
	model.APIKey = types.StringValue(serviceAccount.APIKey.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
