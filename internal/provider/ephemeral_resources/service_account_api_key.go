package ephemeral_resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ ephemeral.EphemeralResource = (*serviceAccountAPIKey)(nil)
)

func NewServiceAccountEphemeralResource() ephemeral.EphemeralResource {
	return &serviceAccountAPIKey{}
}

type serviceAccountAPIKey struct {
	client api.PrefectClient
}

type serviceAccountAPIKeyModel struct {
	AccountID        customtypes.UUIDValue `tfsdk:"account_id"`
	ServiceAccountID customtypes.UUIDValue `tfsdk:"service_account_id"`
	Value            types.String          `tfsdk:"value"`
}

func (k *serviceAccountAPIKey) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_api_key"
}

func (k *serviceAccountAPIKey) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Service account API key",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "The ID of the account to create an API key for.",
				Optional:    true,
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
			},
			"service_account_id": schema.StringAttribute{
				Description: "The ID of the service account to create an API key for.",
				Required:    true,
				CustomType:  customtypes.UUIDType{},
			},
			"value": schema.StringAttribute{
				Description: "The API key value.",
				Computed:    true,
			},
		},
	}
}

func (k *serviceAccountAPIKey) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("ephemeral resource", req.ProviderData))

		return
	}

	k.client = client
}

func (k *serviceAccountAPIKey) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data serviceAccountAPIKeyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := k.client.ServiceAccounts(data.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create service account client", err.Error())

		return
	}

	serviceAccount, err := client.Get(ctx, data.ServiceAccountID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get service account", err.Error())

		return
	}

	// The key is only returned on create/update, so this will always be empty,
	// meaning this won't work.
	data.Value = types.StringValue(serviceAccount.APIKey.Key)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
