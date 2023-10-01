package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var (
	_ = resource.ResourceWithConfigure(&ServiceAccountResource{})
	_ = resource.ResourceWithImportState(&ServiceAccountResource{})
)

type ServiceAccountResource struct {
	client api.PrefectClient
}

type ServiceAccountResourceModel struct {
	ID             types.String               `tfsdk:"id"`
	Name           types.String               `tfsdk:"name"`
	Created        customtypes.TimestampValue `tfsdk:"created"`
	Updated        customtypes.TimestampValue `tfsdk:"updated"`
	AccountID      customtypes.UUIDValue      `tfsdk:"account_id"`

	AccountRoleID  types.String				  `tfsdk:"account_role_id"`
	AccountRoleName       types.String               `tfsdk:"account_role_name"`
	APIKeyID       types.String               `tfsdk:"api_key_id"`
	APIKeyName     types.String               `tfsdk:"api_key_name"`
	APIKeyCreated  customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpires  customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKey         types.String               `tfsdk:"api_key"`
}

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
				Computed: true,
				Description: "Service account UUID",
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Name of the service account",
			},
			"created": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time of the service account creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time that the service account was last updated in RFC 3339 format",
			},
			"account_id": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.UUIDType{},
				Description: "Account UUID, defaults to the account set in the provider",
			},
			// @TODO: Do we actually need 'account_role_id' defined as an attribute here?
			// or just only one of account_role_id or account_role_name?
			"account_role_id": schema.StringAttribute{
				Computed: true,
				Description: "Account Role ID of the service account",
				Optional:    true,
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
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Computed: true,
				Description: "API Key associated with the service account",
			},
		},
	}
}

// Function that copies api.ServiceAccount to a ServiceAccountResourceModel.
func copyServiceAccountToModel(_ context.Context, sa *api.ServiceAccount, model *ServiceAccountResourceModel) diag.Diagnostics {

	model.ID = types.StringValue(sa.ID.String())
	model.Name = types.StringValue(sa.Name)
	model.Created = customtypes.NewTimestampPointerValue(sa.Created)
	model.Updated = customtypes.NewTimestampPointerValue(sa.Updated)
	model.AccountID = customtypes.NewUUIDValue(sa.AccountId)

	model.AccountRoleName = types.StringValue(sa.AccountRoleName)
	model.APIKeyID = types.StringValue(sa.APIKey.Id)
	model.APIKeyName = types.StringValue(sa.APIKey.Name)
	model.APIKeyCreated = customtypes.NewTimestampPointerValue(sa.APIKey.Created)
	model.APIKeyExpires = customtypes.NewTimestampPointerValue(sa.APIKey.Expiration)
	model.APIKey = types.StringValue(sa.APIKey.Key)

	return nil
}

// @TODO later: May need to create a client for AccountRole endpoint and use that to fetch the name using the ID

// 'Create' creates the resource and sets the initial Terraform state.
func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var model ServiceAccountResourceModel

    // Populate the model from resource configuration and emit diagnostics on error
    resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
    if resp.Diagnostics.HasError() {
        return
    }

    client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating Service Account client",
            fmt.Sprintf("Could not create Service Account client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
        )
        return
    }

    // Create a ServiceAccountCreateRequest
    createReq := api.ServiceAccountCreateRequest{
        Name: model.Name.ValueString(),
    }

    // Checking if the timestamp is not empty
	if !model.APIKeyExpires.ValueTime().IsZero() {
		expiration := model.APIKeyExpires.ValueTime().Format(time.RFC3339)
		createReq.APIKeyExpiration = expiration
	}

    // Conditionally set AccountRoleId if it's provided
    if !model.AccountRoleID.IsNull() && !model.AccountRoleID.IsUnknown() {
        createReq.AccountRoleId = model.AccountRoleID.ValueString()
    }

    sa, err := client.Create(ctx, createReq)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating service account",
            fmt.Sprintf("Could not create service account, unexpected error: %s", err),
        )
        return
    }

	// @TODO: May need to pass in / assign the AccountRoleId to the model (required?)
	// Because the response payload for Create does not contain an AccountRoleId
	resp.Diagnostics.Append(copyServiceAccountToModel(ctx, sa, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.AddError(
			"Error creating Service Account client",
			fmt.Sprintf("Could not create Service Account client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	sa, err := client.Get(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyServiceAccountToModel(ctx, sa, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}


// Update updates the resource and sets the updated Terraform state on success.
func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model ServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.ServiceAccounts(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Service Account client",
			fmt.Sprintf("Could not create Service Account client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	// Update client method requires context, botId, request args
	err = client.Update(ctx, model.ID.ValueString(), api.ServiceAccountUpdateRequest{
		Name:      model.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Service Account",
			fmt.Sprintf("Could not update Service Account, unexpected error: %s", err),
		)

		return
	}

	// 'Get' client method requires context, botId args
	sa, err := client.Get(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Service Account state",
			fmt.Sprintf("Could not read Service Account, unexpected error: %s", err),
		)
		return
	}

	// Update the model with latest service account details (from the Get call above)
	// Note: As with the Create/Read operations, 'account role id' is not part of the response
	// and does not get set in the model as part of this call to copyServiceAccountToModel
	resp.Diagnostics.Append(copyServiceAccountToModel(ctx, sa, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
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
		resp.Diagnostics.AddError(
			"Error creating Service Account client",
			fmt.Sprintf("Could not create Service Account client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

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