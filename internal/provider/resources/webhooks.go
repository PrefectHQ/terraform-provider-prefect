package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

const (
	// webhookUpdateRetryAttempts is the number of attempts to fetch a webhook after update.
	webhookUpdateRetryAttempts = 5
	// webhookUpdateRetryDelay is the base delay between retries when fetching after update.
	webhookUpdateRetryDelay = 200 * time.Millisecond
)

var (
	_ = resource.ResourceWithConfigure(&WebhookResource{})
	_ = resource.ResourceWithImportState(&WebhookResource{})
)

type WebhookResource struct {
	client api.PrefectClient
}

type WebhookResourceModel struct {
	ID               customtypes.UUIDValue      `tfsdk:"id"`
	Created          customtypes.TimestampValue `tfsdk:"created"`
	Updated          customtypes.TimestampValue `tfsdk:"updated"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	Enabled          types.Bool                 `tfsdk:"enabled"`
	Template         types.String               `tfsdk:"template"`
	AccountID        customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID      customtypes.UUIDValue      `tfsdk:"workspace_id"`
	Endpoint         types.String               `tfsdk:"endpoint"`
	ServiceAccountID customtypes.UUIDValue      `tfsdk:"service_account_id"`
}

// NewWebhookResource returns a new WebhookResource.
//
//nolint:ireturn // required by Terraform API
func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans("The resource `webhook` represents a Prefect Cloud Webhook. "+
			"Webhooks allow external services to trigger events in Prefect. "+
			"For more information, see [receive events with webhooks](https://docs.prefect.io/v3/automate/events/webhook-triggers).",
			helpers.AllCloudPlans...,
		),
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Webhook ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the webhook",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Description of the webhook",
				Default:     stringdefault.StaticString(""),
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the webhook is enabled",
				Default:     booldefault.StaticBool(true),
			},
			"template": schema.StringAttribute{
				Required:    true,
				Description: "Template used by the webhook. Use jsonencode() for static values or template strings for dynamic values.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was created (RFC3339)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of when the resource was updated (RFC3339)",
			},
			"account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
			},
			"workspace_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
			},
			"endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "The fully-formed webhook endpoint, eg. `https://api.prefect.cloud/hooks/<slug>`",
			},
			"service_account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "ID of the Service Account to which this webhook belongs. `Pro` and `Enterprise` customers can assign a Service Account to a webhook to enhance security. If set, the webhook request will be authorized with the Service Account's API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// copyWebhookResponseToModel maps an API response to a model that is saved in Terraform state.
func copyWebhookResponseToModel(webhook *api.Webhook, tfModel *WebhookResourceModel, endpointHost string) {
	tfModel.ID = customtypes.NewUUIDValue(webhook.ID)
	tfModel.Created = customtypes.NewTimestampPointerValue(webhook.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(webhook.Updated)
	tfModel.Name = types.StringValue(webhook.Name)
	tfModel.Description = types.StringValue(webhook.Description)
	tfModel.Enabled = types.BoolValue(webhook.Enabled)
	tfModel.Template = types.StringValue(webhook.Template)
	tfModel.AccountID = customtypes.NewUUIDValue(webhook.AccountID)
	tfModel.WorkspaceID = customtypes.NewUUIDValue(webhook.WorkspaceID)
	tfModel.Endpoint = types.StringValue(fmt.Sprintf("%s/hooks/%s", endpointHost, webhook.Slug))
	tfModel.ServiceAccountID = customtypes.NewUUIDPointerValue(webhook.ServiceAccountID)
}

// Create creates the resource and sets the initial Terraform state.
func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhookClient, err := r.client.Webhooks(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Webhook", err))

		return
	}

	createReq := api.WebhookCreateRequest{
		WebhookCore: api.WebhookCore{
			Name:             plan.Name.ValueString(),
			Description:      plan.Description.ValueString(),
			Enabled:          plan.Enabled.ValueBool(),
			Template:         plan.Template.ValueString(),
			ServiceAccountID: plan.ServiceAccountID.ValueUUIDPointer(),
		},
	}

	webhook, err := webhookClient.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", "create", err))

		return
	}

	copyWebhookResponseToModel(webhook, &plan, r.client.GetEndpointHost())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() && state.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	client, err := r.client.Webhooks(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Webhook", err))

		return
	}

	var webhook *api.Webhook
	if !state.ID.IsNull() {
		webhook, err = client.Get(ctx, state.ID.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"ID is unset",
			"Webhook ID must be set to retrieve the resource.",
		)

		return
	}

	if err != nil {
		// If the remote object does not exist, we can remove it from TF state
		// so that the framework can queue up a new Create.
		// https://discuss.hashicorp.com/t/recreate-a-resource-in-a-case-of-manual-deletion/66375/3
		if helpers.Is404Error(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", "get", err))

		return
	}

	copyWebhookResponseToModel(webhook, &state, r.client.GetEndpointHost())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Webhooks(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Webhook", err))

		return
	}

	updateReq := api.WebhookUpdateRequest{
		WebhookCore: api.WebhookCore{
			Name:             plan.Name.ValueString(),
			Description:      plan.Description.ValueString(),
			Enabled:          plan.Enabled.ValueBool(),
			Template:         plan.Template.ValueString(),
			ServiceAccountID: plan.ServiceAccountID.ValueUUIDPointer(),
		},
	}

	err = client.Update(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", "update", err))

		return
	}

	// After update, fetch the webhook with retry logic to handle eventual consistency.
	// The API may need a moment to fully process the update before returning
	// the correct state, especially for the 'enabled' field.
	var webhook *api.Webhook
	retryErr := retry.Do(
		func() error {
			var fetchErr error
			webhook, fetchErr = client.Get(ctx, state.ID.ValueString())
			if fetchErr != nil {
				return fmt.Errorf("failed to fetch webhook: %w", fetchErr)
			}

			// Verify the enabled field matches what we expect
			if webhook.Enabled != plan.Enabled.ValueBool() {
				return fmt.Errorf("webhook enabled field not yet updated (expected %v, got %v)", plan.Enabled.ValueBool(), webhook.Enabled)
			}

			return nil
		},
		retry.Attempts(webhookUpdateRetryAttempts),
		retry.Delay(webhookUpdateRetryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
	)

	if retryErr != nil {
		// If retry failed, try one more time without validation
		webhook, err = client.Get(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", "get", err))

			return
		}
	}

	copyWebhookResponseToModel(webhook, &plan, r.client.GetEndpointHost())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Webhooks(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Webhook", err))

		return
	}

	err = client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Webhook", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
