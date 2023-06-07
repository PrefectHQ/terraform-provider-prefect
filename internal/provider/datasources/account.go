package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = datasource.DataSourceWithConfigure(&AccountDataSource{})

// AccountDataSource contains state for the data source.
type AccountDataSource struct {
	client api.AccountsClient
}

// AccountDataSourceModel defines the Terraform data source model.
type AccountDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`

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

	d.client, _ = client.Accounts()
}

// Schema defines the schema for the data source.
func (d *AccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect Cloud account",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Account UUID",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of the account creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time that the account was last updated in RFC 3339 format",
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
				Description: "Billing email to apply to the account's stripe customer",
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

	accountID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Account ID",
			fmt.Sprintf("Could not parse account ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	account, err := d.client.Get(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing account state",
			fmt.Sprintf("Could not read account, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = types.StringValue(account.ID.String())

	if account.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(account.Created.Format(time.RFC3339))
	}

	if account.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(account.Updated.Format(time.RFC3339))
	}

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
