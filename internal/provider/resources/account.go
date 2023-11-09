package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var (
	_ = resource.ResourceWithConfigure(&AccountResource{})
	_ = resource.ResourceWithImportState(&AccountResource{})
)

// AccountResource contains state for the resource.
type AccountResource struct {
	client api.AccountsClient
}

// AccountResourceModel defines the Terraform resource model.
type AccountResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name                  types.String `tfsdk:"name"`
	Handle                types.String `tfsdk:"handle"`
	Location              types.String `tfsdk:"location"`
	Link                  types.String `tfsdk:"link"`
	AllowPublicWorkspaces types.Bool   `tfsdk:"allow_public_workspaces"`
	BillingEmail          types.String `tfsdk:"billing_email"`
}

// NewAccountResource returns a new AccountResource.
//
//nolint:ireturn // required by Terraform API
func NewAccountResource() resource.Resource {
	return &AccountResource{}
}

// Metadata returns the resource type name.
func (r *AccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

// Configure initializes runtime state for the resource.
func (r *AccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client, _ = client.Accounts()
}

// Schema defines the schema for the resource.
func (r *AccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `account` represents a Prefect Cloud account. " +
			"It is used to manage the account's attributes, such as the name, handle, and location.\n" +
			"\n" +
			"Note that this resource can only be imported, as account creation is not currently supported" +
			"via the API. Additionally, be aware that account deletion is possible once it is imported," +
			"so be attentive to any destroy plans or unlink the resource through `terraform state rm`.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				// We cannot use a CustomType due to a conflict with PlanModifiers; see
				// https://github.com/hashicorp/terraform-plugin-framework/issues/763
				// https://github.com/hashicorp/terraform-plugin-framework/issues/754
				Description: "Account ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Description: "Name of the account",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "Unique handle of the account",
				Required:    true,
			},
			"location": schema.StringAttribute{
				Description: "An optional physical location for the account, e.g. Washington, D.C.",
				Required:    true,
			},
			"link": schema.StringAttribute{
				Description: "An optional for an external url associated with the account, e.g. https://prefect.io/",
				Required:    true,
			},
			"allow_public_workspaces": schema.BoolAttribute{
				Description: "Whether or not this account allows public workspaces",
				Required:    true,
			},
			"billing_email": schema.StringAttribute{
				Description: "Billing email to apply to the account's Stripe customer",
				Required:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *AccountResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot create account", "Account is an import-only resource and cannot be created by Terraform.")
}

// copyAccountToModel copies an api.AccountResponse to an AccountResourceModel.
func copyAccountToModel(_ context.Context, account *api.AccountResponse, model *AccountResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(account.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(account.Created)
	model.Updated = customtypes.NewTimestampPointerValue(account.Updated)

	model.AllowPublicWorkspaces = types.BoolPointerValue(account.AllowPublicWorkspaces)
	model.BillingEmail = types.StringPointerValue(account.BillingEmail)
	model.Handle = types.StringValue(account.Handle)
	model.Link = types.StringPointerValue(account.Link)
	model.Location = types.StringPointerValue(account.Location)
	model.Name = types.StringValue(account.Name)

	return nil
}

// Read refreshes the Terraform state with the latest data.
func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model AccountResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
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

	account, err := r.client.Get(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing account state",
			fmt.Sprintf("Could not read account, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyAccountToModel(ctx, account, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model AccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
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

	err = r.client.Update(ctx, accountID, api.AccountUpdate{
		Name:                  model.Name.ValueStringPointer(),
		Handle:                model.Handle.ValueStringPointer(),
		Location:              model.Location.ValueStringPointer(),
		Link:                  model.Link.ValueStringPointer(),
		AllowPublicWorkspaces: model.AllowPublicWorkspaces.ValueBoolPointer(),
		BillingEmail:          model.BillingEmail.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating account",
			fmt.Sprintf("Could not update account, unexpected error: %s", err),
		)

		return
	}

	account, err := r.client.Get(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing account state",
			fmt.Sprintf("Could not read account, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyAccountToModel(ctx, account, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model AccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
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

	err = r.client.Delete(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting account",
			fmt.Sprintf("Could not delete account, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
