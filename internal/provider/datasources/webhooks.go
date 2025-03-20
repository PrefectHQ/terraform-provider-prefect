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

var _ = datasource.DataSourceWithConfigure(&WebhookDataSource{})

// WebhookDataSource contains state for the data source.
type WebhookDataSource struct {
	client api.PrefectClient
}

// WebhookDataSourceModel defines the Terraform data source model.
type WebhookDataSourceModel struct {
	ID               customtypes.UUIDValue      `tfsdk:"id"`
	Created          customtypes.TimestampValue `tfsdk:"created"`
	Updated          customtypes.TimestampValue `tfsdk:"updated"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	Enabled          types.Bool                 `tfsdk:"enabled"`
	Template         types.String               `tfsdk:"template"`
	AccountID        customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID      customtypes.UUIDValue      `tfsdk:"workspace_id"`
	Slug             types.String               `tfsdk:"slug"`
	ServiceAccountID customtypes.UUIDValue      `tfsdk:"service_account_id"`
}

// NewWebhookDataSource returns a new WebhookDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWebhookDataSource() datasource.DataSource {
	return &WebhookDataSource{}
}

// Metadata returns the data source type name.
func (d *WebhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

// Configure initializes runtime state for the data source.
func (d *WebhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}

var webhookAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Webhook ID (UUID)",
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
		Optional:    true,
		Description: "Name of the webhook",
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of the webhook",
	},
	"enabled": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the webhook is enabled",
	},
	"template": schema.StringAttribute{
		Computed:    true,
		Description: "Template used by the webhook",
	},
	"account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID), defaults to the account set in the provider",
		Optional:    true,
	},
	"workspace_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
		Optional:    true,
	},
	"slug": schema.StringAttribute{
		Computed:    true,
		Description: "Slug of the webhook",
	},
	"service_account_id": schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "ID of the Service Account to which this webhook belongs. `Pro` and `Enterprise` customers can assign a Service Account to a webhook to enhance security.",
		Computed:    true,
	},
}

// Schema defines the schema for the data source.
func (d *WebhookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Webhook, by name or ID.
<br>
Use this data source to obtain webhook-level attributes, such as ID, Name, Template, and more.
<br>
For more information, see [receive events with webhooks](https://docs.prefect.io/v3/automate/events/webhook-triggers).
`,
			helpers.AllCloudPlans...,
		),
		Attributes: webhookAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WebhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WebhookDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() && model.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"Either a Webhook ID or Name is required to read a Webhook.",
		)

		return
	}

	client, err := d.client.Webhooks(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Webhook", err))

		return
	}

	// A Webhook can be read by either ID or Name.
	// If both are set, we prefer the ID
	var webhook *api.Webhook
	var operation string
	if !model.ID.IsNull() {
		operation = "get"
		webhook, err = client.Get(ctx, model.ID.ValueString())
	} else if !model.Name.IsNull() {
		var webhooks []*api.Webhook
		operation = "list"
		webhooks, err = client.List(ctx, []string{model.Name.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific service account is not returned
		if err == nil && len(webhooks) != 1 {
			err = fmt.Errorf("a Webhook with the name=%s could not be found", model.Name.ValueString())
		}

		if len(webhooks) == 1 {
			webhook = webhooks[0]
		}
	}

	if webhook == nil {
		resp.Diagnostics.AddError(
			"Error refreshing Webhook state",
			fmt.Sprintf("Could not find Webhook with ID=%s and Name=%s", model.ID.ValueString(), model.Name.ValueString()),
		)

		return
	}

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", operation, err))

		return
	}

	model.ID = customtypes.NewUUIDValue(webhook.ID)
	model.Created = customtypes.NewTimestampPointerValue(webhook.Created)
	model.Updated = customtypes.NewTimestampPointerValue(webhook.Updated)

	model.Name = types.StringValue(webhook.Name)
	model.Description = types.StringValue(webhook.Description)
	model.Enabled = types.BoolValue(webhook.Enabled)
	model.Template = types.StringValue(webhook.Template)
	model.AccountID = customtypes.NewUUIDValue(webhook.AccountID)
	model.WorkspaceID = customtypes.NewUUIDValue(webhook.WorkspaceID)
	model.Slug = types.StringValue(webhook.Slug)
	model.ServiceAccountID = customtypes.NewUUIDPointerValue(webhook.ServiceAccountID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
