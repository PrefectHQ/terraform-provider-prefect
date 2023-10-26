package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&ServiceAccountResource{})
	_ = resource.ResourceWithImportState(&ServiceAccountResource{})
)

type ServiceAccountResource struct {
	client api.PrefectClient
}

type ServiceAccountResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name            types.String          `tfsdk:"name"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	AccountRoleName types.String          `tfsdk:"account_role_name"`

	APIKeyID         types.String               `tfsdk:"api_key_id"`
	APIKeyName       types.String               `tfsdk:"api_key_name"`
	APIKeyCreated    customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpiration customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKey           types.String               `tfsdk:"api_key"`
}

// ArePointerTimesEqual is a helper to compare equality of two pointer times
// as this can get verbose to do inline with the resource logic.
func ArePointerTimesEqual(t1 *time.Time, t2 *time.Time) bool {
	return t1 == t2 || (t1 != nil && t2 != nil && t1.Equal(*t2))
}

// NewServiceAccountResource returns a new AccountResource.
//
//nolint:ireturn // required by Terraform API
func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource representing a Prefect service account",
		Version:     1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service account UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the service account",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Date and time of the service account creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Date and time that the service account was last updated in RFC 3339 format",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account UUID, defaults to the account set in the provider",
			},
			"account_role_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Account Role name of the service account",
			},
			"api_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "API Key ID associated with the service account",
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
				Optional:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Date and time that the API Key expires in RFC 3339 format",
			},
			"api_key": schema.StringAttribute{
				Computed:    true,
				Description: "API Key associated with the service account",
				Sensitive:   true,
			},
		},
	}
}

// Function that copies api.ServiceAccount to a ServiceAccountResourceModel.
// NOTE: the API Key is attached to the resource model outside of this helper,
// as it is only returned on Create/Update operations.
func copyServiceAccountResponseToModel(serviceAccount *api.ServiceAccount, model *ServiceAccountResourceModel) {
	model.ID = types.StringValue(serviceAccount.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(serviceAccount.Created)
	model.Updated = customtypes.NewTimestampPointerValue(serviceAccount.Updated)

	model.Name = types.StringValue(serviceAccount.Name)
	model.AccountID = customtypes.NewUUIDValue(serviceAccount.AccountID)
	model.AccountRoleName = types.StringValue(serviceAccount.AccountRoleName)

	model.APIKeyID = types.StringValue(serviceAccount.APIKey.ID)
	model.APIKeyName = types.StringValue(serviceAccount.APIKey.Name)
	model.APIKeyCreated = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Created)
	model.APIKeyExpiration = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Expiration)
}

// Create creates the resource and sets the initial Terraform state.
func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	createReq := api.ServiceAccountCreateRequest{
		Name: model.Name.ValueString(),
	}

	// Conditionally set APIKeyExpiration if it's provided
	if !model.APIKeyExpiration.ValueTime().IsZero() {
		expiration := model.APIKeyExpiration.ValueTime().Format(time.RFC3339)
		createReq.APIKeyExpiration = expiration
	}

	// @TODO
	// Using the account_roles client, fetch the account role ID using the name
	// that is passed into the service account resource
	// https://github.com/PrefectHQ/terraform-provider-prefect/issues/39
	// createReq.AccountRoleID = "0000-0000-0000-0000-0000"

	serviceAccount, err := client.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service account",
			fmt.Sprintf("Could not create service account, unexpected error: %s", err),
		)

		return
	}

	copyServiceAccountResponseToModel(serviceAccount, &model)

	// The API Key is only returned on Create or when rotating the key, so we'll attach it to
	// the model outside of the helper function, so that we can prevent the value from being
	// overwritten in state when this helper is used on Read operations.
	model.APIKey = types.StringValue(serviceAccount.APIKey.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	serviceAccount, err := client.Get(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err),
		)

		return
	}

	copyServiceAccountResponseToModel(serviceAccount, &model)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	// Update client method requires context, botID, request args
	err = client.Update(ctx, plan.ID.ValueString(), api.ServiceAccountUpdateRequest{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Service Account",
			fmt.Sprintf("Could not update Service Account, unexpected error: %s", err),
		)

		return
	}

	serviceAccount, err := client.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err),
		)

		return
	}

	// Here, we'll retrieve the API Key from the previous State, as it's
	// not included in the Terraform Plan / user configuration, nor is it
	// returned on any API response other than Create and RotateKey.
	// Additionally, the Provider framework will throw an exception if we
	// set the `api_key` property to use stringmodifier.UseStateForUnknown()
	// during legitimate cases where the API Key State value will be updated
	// during key rotation.
	var state ServiceAccountResourceModel
	req.State.Get(ctx, &state)
	apiKey := state.APIKey.ValueString()

	// Practitioners can rotate their Service Account API Key my modifying the
	// `api_key_expiration` attribute. If the provided value is different than the current
	// value, we'll call the RotateKey method on the client, which returns the
	// ServiceAccount object with the new API Key value included in the response.
	providedExpiration := plan.APIKeyExpiration.ValueTimePointer()
	currentExpiration := serviceAccount.APIKey.Expiration
	if !ArePointerTimesEqual(providedExpiration, currentExpiration) {
		serviceAccount, err = client.RotateKey(ctx, plan.ID.ValueString(), api.ServiceAccountRotateKeyRequest{
			APIKeyExpiration: providedExpiration,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error rotating Service Account key",
				fmt.Sprintf("Could not rotate Service Account key, unexpected error: %s", err),
			)

			return
		}

		// Upon successful key rotation, we'll update this local variable with the new API Key value,
		// which will be used in the final State update below.
		apiKey = serviceAccount.APIKey.Key
	}

	// Update the model with latest service account details (from the Get call above)
	copyServiceAccountResponseToModel(serviceAccount, &plan)

	// The API Key is only returned on Create or when rotating the key, so we'll attach it to
	// the model outside of the helper function, so that we can prevent the value from being
	// overwritten in state when this helper is used on Read operations.
	plan.APIKey = types.StringValue(apiKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	err = client.Delete(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Service Account",
			fmt.Sprintf("Could not delete Service Account, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
