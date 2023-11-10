package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = datasource.DataSourceWithConfigure(&AccountDataSource{})

// AccountDataSource contains state for the data source.
type AccountDataSource struct {
	client api.PrefectClient
}

// AccountDataSourceModel defines the Terraform data source model.
type AccountDataSourceModel struct {
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name                  types.String `tfsdk:"name"`
	Handle                types.String `tfsdk:"handle"`
	Location              types.String `tfsdk:"location"`
	Link                  types.String `tfsdk:"link"`
	AllowPublicWorkspaces types.Bool   `tfsdk:"allow_public_workspaces"`
	BillingEmail          types.String `tfsdk:"billing_email"`
}

// NewAccountDataSource returns a new AccountDataSource.
//
//nolint:ireturn // required by Terraform API
func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

// Metadata returns the data source type name.
func (d *AccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

// Configure initializes runtime state for the data source.
func (d *AccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema defines the schema for the data source.
func (d *AccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Description: "Data Source representing a Prefect Cloud account",
		Description: `
Get information about an existing Account.
<br>
Use this data source to obtain account-level attributes
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID)",
				Optional:    true,
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
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the account",
			},
			"handle": schema.StringAttribute{
				Computed:    true,
				Description: "Unique handle of the account",
			},
			"location": schema.StringAttribute{
				Computed:    true,
				Description: "An optional physical location for the account, e.g. Washington, D.C.",
			},
			"link": schema.StringAttribute{
				Computed:    true,
				Description: "An optional for an external url associated with the account, e.g. https://prefect.io/",
			},
			"allow_public_workspaces": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether or not this account allows public workspaces",
			},
			"billing_email": schema.StringAttribute{
				Computed:    true,
				Description: "Billing email to apply to the account's Stripe customer",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *AccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model AccountDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The ID value will be validated at the Schema level,
	// so we can safely convert + extract the UUID value here
	// without checking for an error. If a null value is passed,
	// we'll fall back to the account_id set in the Accounts client.
	accountID := model.ID.ValueUUID()

	client, err := d.client.Accounts(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	account, err := client.Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing account state",
			fmt.Sprintf("Could not read account, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(account.ID)
	model.Created = customtypes.NewTimestampPointerValue(account.Created)
	model.Updated = customtypes.NewTimestampPointerValue(account.Updated)

	model.AllowPublicWorkspaces = types.BoolPointerValue(account.AllowPublicWorkspaces)
	model.BillingEmail = types.StringPointerValue(account.BillingEmail)
	model.Handle = types.StringValue(account.Handle)
	model.Link = types.StringPointerValue(account.Link)
	model.Location = types.StringPointerValue(account.Location)
	model.Name = types.StringValue(account.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
