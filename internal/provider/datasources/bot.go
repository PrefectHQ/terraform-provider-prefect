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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &BotDataSource{}
	_ datasource.DataSourceWithConfigure = &BotDataSource{}
)

// BotDataSource contains state for the data source modeling a Prefect Service Account.
type BotDataSource struct {
	client api.PrefectClient
}

// AccountDataSourceModel defines the Terraform data source model.
// the TF data source configuration will be unmarsheled into this struct
// NOTE: the APIKey field is not included in bot fetches and
// is excluded from this datasource model.
type BotDataSourceModel struct {
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name            types.String          `tfsdk:"name"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	AccountRoleName types.String          `tfsdk:"account_role_name"`
}

// NewBotDataSource returns a new BotDataSource,
// to be inserted into the provider during instantiation.
//
//nolint:ireturn // required by Terraform API
func NewBotDataSource() datasource.DataSource {
	return &BotDataSource{}
}

// Metadata returns the data source type name.
func (d *BotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bot"
}

func (d *BotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Source representing a Prefect Service Account",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Service Account UUID",
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
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Service Account",
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account UUID where Service Account resides",
			},
			"account_role_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Service Account",
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *BotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model BotDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Bots()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating the Bots client",
			fmt.Sprintf("Could not create Bots client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	bot, err := client.Get(ctx, model.ID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to fetch Bot and refresh state",
			fmt.Sprintf("Could not fetch Bot, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(bot.ID)
	model.Created = customtypes.NewTimestampPointerValue(bot.Created)
	model.Updated = customtypes.NewTimestampPointerValue(bot.Updated)

	model.Name = types.StringValue(bot.Name)
	model.AccountID = customtypes.NewUUIDValue(bot.AccountID)
	model.AccountRoleName = types.StringValue(bot.AccountRoleName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
