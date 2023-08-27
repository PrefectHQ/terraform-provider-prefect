package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var _ = datasource.DataSourceWithConfigure(&ServiceAccountDataSource{})

// ServiceAccountDataSource contains state for the data source.
type ServiceAccountDataSource struct {
	client api.PrefectClient
}

// ServiceAccountSourceModel defines the Terraform data source model.
type ServiceAccountSourceModel struct {
	ID          customtypes.UUIDValue      `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`

	// SA fields
	AccountRoleName       types.String        `tfsdk:"account_role_name"`
	APIKeyID       types.String               `tfsdk:"api_key_id"`
	APIKeyName     types.String               `tfsdk:"api_key_name"`
	APIKeyCreated  customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpires  customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKey         types.String               `tfsdk:"api_key"`
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
		CustomType:  customtypes.UUIDType{},
		Description: "Service Account UUID",
		Optional:    true,
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the Service Account creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the Service Account was last updated in RFC 3339 format",
	},
	"account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account UUID, defaults to the account set in the provider",
		Optional:    true,
	},
	"name": schema.StringAttribute{
		Required: true,
		Description: "Name of the service account",
	},
	"account_role_name": schema.StringAttribute{
		Computed: true,
		Description: "Account Role name of the service account",
	},
	"api_key_id": schema.StringAttribute{
		Computed: true,
		Description: "API Key ID associated with the service account",
	},
	"api_key_name": schema.StringAttribute{
		Computed: true,
		Description: "API Key Name associated with the service account",
	},
	"api_key_created": schema.StringAttribute{
		Computed: true,
		CustomType: customtypes.TimestampType{},
		Description: "Date and time that the API Key was created in RFC 3339 format",
	},
	"api_key_expiration": schema.StringAttribute{
		Computed: true,
		CustomType: customtypes.TimestampType{},
		Description: "Date and time that the API Key expires in RFC 3339 format",
	},
	"api_key": schema.StringAttribute{
		Computed: true,
		Description: "API Key associated with the service account",
	},
}

// Schema defines the schema for the data source.
func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect Service Account",
		Attributes:  serviceAccountAttributes,
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

	client, err := d.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable client",
			fmt.Sprintf("Could not create variable client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	sa, err := client.Get(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(sa.ID)
	model.Created = customtypes.NewTimestampPointerValue(sa.Created)
	model.Updated = customtypes.NewTimestampPointerValue(sa.Updated)
	model.AccountID = customtypes.NewUUIDValue(sa.AccountID)

	model.AccountRoleName = types.StringValue(sa.AccountRoleName)
	model.APIKeyID = types.StringValue(sa.APIKey.Id)
	model.APIKeyName = types.StringValue(sa.APIKey.Name)
	model.APIKeyCreated = customtypes.NewTimestampPointerValue(sa.APIKey.Created)
	model.APIKeyExpires = customtypes.NewTimestampPointerValue(sa.APIKey.Expiration)
	model.APIKey = types.StringValue(sa.APIKey.Key)

	if err != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
