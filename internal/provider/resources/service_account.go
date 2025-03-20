package resources

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

const (
	oldKeyExpiresInSecondsMin = 0
	oldKeyExpiresInSecondsMax = 172800
)

var (
	_ = resource.ResourceWithConfigure(&ServiceAccountResource{})
	_ = resource.ResourceWithImportState(&ServiceAccountResource{})
)

type ServiceAccountResource struct {
	client api.PrefectClient
}

type ServiceAccountResourceModel struct {
	BaseModel

	Name            types.String          `tfsdk:"name"`
	ActorID         customtypes.UUIDValue `tfsdk:"actor_id"`
	AccountID       customtypes.UUIDValue `tfsdk:"account_id"`
	AccountRoleName types.String          `tfsdk:"account_role_name"`

	APIKeyID               types.String               `tfsdk:"api_key_id"`
	APIKeyName             types.String               `tfsdk:"api_key_name"`
	APIKeyCreated          customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpiration       customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKeyKeepers          types.Map                  `tfsdk:"api_key_keepers"`
	OldKeyExpiresInSeconds types.Int32                `tfsdk:"old_key_expires_in_seconds"`
	APIKey                 types.String               `tfsdk:"api_key"`
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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `service_account` represents a Prefect Cloud Service Account. "+
				"A Service Account allows you to create an API Key that is not associated with a user account.\n"+
				"\n"+
				"Service Accounts are used to configure API access for workers or programs. Use this resource to provision "+
				"and rotate Keys as well as assign Account and Workspace Access through Roles.\n"+
				"\n"+
				"API Keys for `service_account` resources can be rotated by modifying the `api_key_expiration` attribute.\n"+
				"For more information, see [manage service accounts](https://docs.prefect.io/v3/manage/cloud/manage-users/service-accounts).",
			helpers.PlanPrefectCloudPro,
			helpers.PlanPrefectCloudEnterprise,
		),
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Service account ID (UUID)",
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
			"actor_id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Actor ID (UUID), used for granting access to resources like Blocks and Deployments",
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
			},
			"account_role_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Account Role name of the service account (one of: Admin, Member, Owner)",
				Default:     stringdefault.StaticString("Member"),
				Validators: []validator.String{
					stringvalidator.OneOf("Admin", "Member", "Owner"),
				},
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
				Description: "Timestamp of the API Key creation (RFC3339)",
			},
			"api_key_expiration": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Timestamp of the API Key expiration (RFC3339). If left as null, the API Key will not expire. Modify this attribute to force a key rotation.",
			},
			"api_key_keepers": schema.MapAttribute{
				Optional:    true,
				Description: "A map of values that, if changed, will trigger a key rotation (but not a re-creation of the Service Account)",
				ElementType: types.StringType,
			},
			"old_key_expires_in_seconds": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Provide this field to set an expiration for the currently active api key. If not provided or provided Null, the current key will be deleted. If provided, it cannot be more than 48 hours (172800 seconds) in the future.",
				Default:     int32default.StaticInt32(oldKeyExpiresInSecondsMin),
				Validators: []validator.Int32{
					int32validator.AtMost(oldKeyExpiresInSecondsMax),
					int32validator.AtLeast(oldKeyExpiresInSecondsMin),
				},
			},
			"api_key": schema.StringAttribute{
				Computed:    true,
				Description: "API Key associated with the service account",
				Sensitive:   true,
			},
		},
	}
}

// copyServiceAccountToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyServiceAccountToModel(serviceAccount *api.ServiceAccount, tfModel *ServiceAccountResourceModel) {
	// NOTE: the API Key is attached to the resource model outside of this helper,
	// as it is only returned on Create/Update operations.
	tfModel.ID = types.StringValue(serviceAccount.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(serviceAccount.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(serviceAccount.Updated)

	tfModel.Name = types.StringValue(serviceAccount.Name)
	tfModel.ActorID = customtypes.NewUUIDValue(serviceAccount.ActorID)
	tfModel.AccountID = customtypes.NewUUIDValue(serviceAccount.AccountID)
	tfModel.AccountRoleName = types.StringValue(serviceAccount.AccountRoleName)

	tfModel.APIKeyID = types.StringValue(serviceAccount.APIKey.ID)
	tfModel.APIKeyName = types.StringValue(serviceAccount.APIKey.Name)
	tfModel.APIKeyCreated = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Created)
	tfModel.APIKeyExpiration = customtypes.NewTimestampPointerValue(serviceAccount.APIKey.Expiration)
}

// Create creates the resource and sets the initial Terraform state.
func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountClient, err := r.client.ServiceAccounts(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	createReq := api.ServiceAccountCreateRequest{
		Name: plan.Name.ValueString(),
	}

	// If the Account Role Name is provided, we'll need to fetch the Account Role ID
	// and attach it to the Create request.
	if !plan.AccountRoleName.IsNull() && !plan.AccountRoleName.IsUnknown() {
		accountRoleClient, err := r.client.AccountRoles(plan.AccountID.ValueUUID())
		if err != nil {
			resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account Role", err))

			return
		}

		accountRoles, err := accountRoleClient.List(ctx, []string{plan.AccountRoleName.ValueString()})
		if err != nil {
			resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account Role", "list", err))

			return
		}

		if len(accountRoles) != 1 {
			resp.Diagnostics.AddError(
				"Could not find Account Role",
				fmt.Sprintf("Could not find Account Role with name %s", plan.AccountRoleName.String()),
			)

			return
		}

		createReq.AccountRoleID = &accountRoles[0].ID
	}

	// Conditionally set APIKeyExpiration if it's provided
	if !plan.APIKeyExpiration.ValueTime().IsZero() {
		expiration := plan.APIKeyExpiration.ValueTime().Format(time.RFC3339)
		createReq.APIKeyExpiration = expiration
	}

	serviceAccount, err := serviceAccountClient.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", "create", err))

		return
	}

	copyServiceAccountToModel(serviceAccount, &plan)

	// The API Key is only returned on Create or when rotating the key, so we'll attach it to
	// the model outside of the helper function, so that we can prevent the value from being
	// overwritten in state when this helper is used on Read operations.
	plan.APIKey = types.StringValue(serviceAccount.APIKey.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountResourceModel

	// Populate the model from state and emit diagnostics on error
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

	client, err := r.client.ServiceAccounts(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	// A Service Account can be read by either ID or Name.
	// If both are set, we prefer the ID
	var serviceAccount *api.ServiceAccount
	var operation string
	if !state.ID.IsNull() {
		operation = "get"
		serviceAccount, err = client.Get(ctx, state.ID.ValueString())
	} else if !state.Name.IsNull() {
		var serviceAccounts []*api.ServiceAccount
		operation = "list"
		serviceAccounts, err = client.List(ctx, []string{state.Name.ValueString()})

		// The error from the API call should take precedence
		// followed by this custom error if a specific service account is not returned
		if err == nil && len(serviceAccounts) != 1 {
			err = fmt.Errorf("a Service Account with the name=%s could not be found", state.Name.ValueString())
		}

		if len(serviceAccounts) == 1 {
			serviceAccount = serviceAccounts[0]
		}
	}

	if serviceAccount == nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not find Service Account with ID=%s and Name=%s", state.ID.ValueString(), state.Name.ValueString()),
		)

		return
	}

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", operation, err))

		return
	}

	copyServiceAccountToModel(serviceAccount, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state ServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(plan.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	// Here, we'll retrieve the API Key from the previous State, as it's
	// not included in the Terraform Plan / user configuration, nor is it
	// returned on any API response other than Create and RotateKey.
	// This means that we'll want to grab + persist the value in State
	// if no rotation takes place. If a rotation does take place, we'll
	// update this variable with the new API Key value returned from that call,
	// and persist that into State.
	// Note that using the stringplanmodifier.UseStateForUnknown() helper for
	// this attribute will not work in this case, as the Provider will throw an
	// error upon key rotation, as the value coming back will be different with
	// the previous value held in State.
	apiKey := state.APIKey.ValueString()

	updateReq := api.ServiceAccountUpdateRequest{
		Name: plan.Name.ValueString(),
	}

	// Check if the Account Role Name was specifically changed in the plan,
	// so we can only perform the Account Role lookup if we need to.
	accountRoleName := state.AccountRoleName.ValueString()
	if accountRoleName != plan.AccountRoleName.ValueString() {
		accountRoleClient, err := r.client.AccountRoles(plan.AccountID.ValueUUID())
		if err != nil {
			resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account Role", err))

			return
		}

		accountRoles, err := accountRoleClient.List(ctx, []string{plan.AccountRoleName.ValueString()})
		if err != nil {
			resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account Role", "list", err))

			return
		}

		if len(accountRoles) != 1 {
			resp.Diagnostics.AddError(
				"Could not find Account Role",
				fmt.Sprintf("Could not find Account Role with name %s", plan.AccountRoleName.String()),
			)

			return
		}

		updateReq.AccountRoleID = &accountRoles[0].ID
	}

	// Update client method requires context, botID, request args
	err = client.Update(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", "update", err))

		return
	}

	serviceAccount, err := client.Get(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", "get", err))

		return
	}

	// Practitioners can rotate their Service Account API Key by modifying the
	// `api_key_expiration` attribute. If the provided value is different than the current
	// value, we'll call the RotateKey method on the client, which returns the
	// ServiceAccount object with the new API Key value included in the response.
	providedExpiration := plan.APIKeyExpiration.ValueTimePointer()
	currentExpiration := serviceAccount.APIKey.Expiration

	// Optionally, practitioners can rotate the key by modifying the `api_key_keepers` map.
	// This is useful for rotating keys that are not expiring.
	currentKeepers := state.APIKeyKeepers.Elements()
	providedKeepers := plan.APIKeyKeepers.Elements()

	if !ArePointerTimesEqual(providedExpiration, currentExpiration) || !maps.Equal(currentKeepers, providedKeepers) {
		serviceAccount, err = client.RotateKey(ctx, plan.ID.ValueString(), api.ServiceAccountRotateKeyRequest{
			APIKeyExpiration:       providedExpiration,
			OldKeyExpiresInSeconds: plan.OldKeyExpiresInSeconds.ValueInt32(),
		})
		if err != nil {
			resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", "key rotate", err))

			return
		}

		// Upon successful key rotation, we'll update this local variable with the new API Key value,
		// which will be used in the final State update below.
		apiKey = serviceAccount.APIKey.Key
	}

	// Update the model with latest service account details (from the Get call above)
	copyServiceAccountToModel(serviceAccount, &plan)

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
	var state ServiceAccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(state.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Service Account", err))

		return
	}

	err = client.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Service Account", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.HasPrefix(req.ID, "name/") {
		name := strings.TrimPrefix(req.ID, "name/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	} else {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	}
}
