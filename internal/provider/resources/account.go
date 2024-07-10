package resources

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&AccountResource{})
	_ = resource.ResourceWithImportState(&AccountResource{})
)

// AccountResource contains state for the resource.
type AccountResource struct {
	client api.PrefectClient
}

// AccountResourceModel defines the Terraform resource model.
type AccountResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	Name         types.String `tfsdk:"name"`
	Handle       types.String `tfsdk:"handle"`
	Location     types.String `tfsdk:"location"`
	Link         types.String `tfsdk:"link"`
	Settings     types.Object `tfsdk:"settings"`
	BillingEmail types.String `tfsdk:"billing_email"`
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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *AccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `account` represents a Prefect Cloud account. " +
			"It is used to manage the account's attributes, such as the name, handle, and location.\n" +
			"\n" +
			"Note that this resource can only be imported, as account creation is not currently supported " +
			"via the API. Additionally, be aware that account deletion is possible once it is imported, " +
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:    true,
			},
			"link": schema.StringAttribute{
				Description: "An optional for an external url associated with the account, e.g. https://prefect.io/",
				Optional:    true,
			},
			"settings": schema.SingleNestedAttribute{
				Description: "Group of settings related to accounts",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"allow_public_workspaces": schema.BoolAttribute{
						Description: "Whether or not this account allows public workspaces",
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ai_log_summaries": schema.BoolAttribute{
						Description: "Whether to use AI to generate log summaries.",
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
					"managed_execution": schema.BoolAttribute{
						Description: "Whether to enable the use of managed work pools",
						Optional:    true,
						Default:     booldefault.StaticBool(true),
					},
				},
			},
			"billing_email": schema.StringAttribute{
				Description: "Billing email to apply to the account's Stripe customer",
				Optional:    true,
			},
		},
	}
}

// copyAccountToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyAccountToModel(_ context.Context, account *api.AccountResponse, tfModel *AccountResourceModel) diag.Diagnostics {
	tfModel.ID = types.StringValue(account.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(account.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(account.Updated)

	tfModel.BillingEmail = types.StringPointerValue(account.BillingEmail)
	tfModel.Handle = types.StringValue(account.Handle)
	tfModel.Link = types.StringPointerValue(account.Link)
	tfModel.Location = types.StringPointerValue(account.Location)
	tfModel.Name = types.StringValue(account.Name)

	settingsObject, diags := types.ObjectValue(
		map[string]attr.Type{
			"allow_public_workspaces": types.BoolType,
			"ai_log_summaries":        types.BoolType,
			"managed_execution":       types.BoolType,
		},
		map[string]attr.Value{
			"allow_public_workspaces": types.BoolValue(*account.Settings.AllowPublicWorkspaces),
			"ai_log_summaries":        types.BoolValue(*account.Settings.AILogSummaries),
			"managed_execution":       types.BoolValue(*account.Settings.ManagedExecution),
		},
	)

	tfModel.Settings = settingsObject

	return diags
}

func (r *AccountResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Cannot create account", "Account is an import-only resource and cannot be created by Terraform.")
}

// Read refreshes the Terraform state with the latest data.
func (r *AccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccountResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.Accounts(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	account, err := client.Get(ctx)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "get", err))

		return
	}

	resp.Diagnostics.Append(copyAccountToModel(ctx, account, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.Accounts(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	settings := plan.Settings.Attributes()

	allowPublicWorkspaces, _ := strconv.ParseBool(settings["allow_public_workspaces"].String())
	aiLogSummaries, _ := strconv.ParseBool(settings["ai_log_summaries"].String())
	managedExecution, _ := strconv.ParseBool(settings["managed_execution"].String())

	err = client.Update(ctx, api.AccountUpdate{
		Name:     plan.Name.ValueStringPointer(),
		Handle:   plan.Handle.ValueStringPointer(),
		Location: plan.Location.ValueStringPointer(),
		Link:     plan.Link.ValueStringPointer(),
		Settings: &api.Settings{
			AllowPublicWorkspaces: &allowPublicWorkspaces,
			AILogSummaries:        &aiLogSummaries,
			ManagedExecution:      &managedExecution,
		},
		BillingEmail: plan.BillingEmail.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "update", err))

		return
	}

	account, err := client.Get(ctx)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "get", err))

		return
	}

	resp.Diagnostics.Append(copyAccountToModel(ctx, account, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccountResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Account", err))

		return
	}

	client, err := r.client.Accounts(accountID)
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Account", err))
	}

	err = client.Delete(ctx)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Account", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *AccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
