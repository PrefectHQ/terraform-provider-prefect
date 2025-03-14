package resources

import (
	"context"

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
	_ = resource.ResourceWithConfigure(&UserAPIKeyResource{})
)

// UserAPIKeyResource contains state for the user API key resource.
type UserAPIKeyResource struct {
	client api.PrefectClient
}

// UserAPIKeyResourceModel defines the Terraform resource model.
type UserAPIKeyResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`

	UserID     types.String               `tfsdk:"user_id"`
	Name       types.String               `tfsdk:"name"`
	Expiration customtypes.TimestampValue `tfsdk:"expiration"`
	Key        types.String               `tfsdk:"key"`
}

// NewUserAPIKeyResource returns a new UserAPIKeyResource.
//
//nolint:ireturn // required by Terraform API
func NewUserAPIKeyResource() resource.Resource {
	return &UserAPIKeyResource{}
}

// Metadata returns the resource type name.
func (r *UserAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_api_key"
}

// Configure initializes runtime state for the resource.
func (r *UserAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Schema returns the resource schema.
func (r *UserAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `user_api_key` represents a Prefect User API Key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "User API Key ID (UUID)",
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
			"user_id": schema.StringAttribute{
				Description: "User ID (UUID)",
				Required:    true,
				// Changing the user ID should force a replacement of the API key
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the API key",
				Required:    true,
				// API key name is only set on create, and
				// we do not support modifying this value. Therefore, any changes
				// to this attribute will force a replacement.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.TimestampType{},
				Description: "Expiration of the API key (RFC3339). If left as null, the API key will not expire. Modify this attribute to re-create the API key.",
				// API key expiration is only set on create, and
				// we do not support modifying this value. Therefore, any changes
				// to this attribute will force a replacement.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "Value of the API key",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// copyUserAPIKeyToModel copies the UserAPIKey resource data to the Terraform model.
// Note: we do not copy the Key field to the model, as it is only returned on Create.
// For all other lifecycle methods, we will persist the existing State value.
func copyUserAPIKeyToModel(apiKey *api.UserAPIKey, model *UserAPIKeyResourceModel) {
	model.ID = types.StringValue(apiKey.ID.String())
	model.Created = customtypes.NewTimestampValue(apiKey.Created)
	model.Name = types.StringValue(apiKey.Name)
	model.Expiration = customtypes.NewTimestampPointerValue(apiKey.Expiration)
}

// Create creates a new User API Key.
func (r *UserAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserAPIKeyResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userClient, err := r.client.Users()
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("User", err))

		return
	}

	createReq := api.UserAPIKeyCreate{
		Name:       plan.Name.ValueString(),
		Expiration: plan.Expiration.ValueTimePointer(),
	}

	apiKey, err := userClient.CreateAPIKey(ctx, plan.UserID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User API Key", "create", err))

		return
	}

	copyUserAPIKeyToModel(apiKey, &plan)
	plan.Key = types.StringValue(apiKey.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *UserAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserAPIKeyResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		resp.Diagnostics.AddError(
			"ID is unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	userClient, err := r.client.Users()
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("User", err))

		return
	}

	apiKey, err := userClient.ReadAPIKey(ctx, state.UserID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User API Key", "read", err))

		return
	}

	copyUserAPIKeyToModel(apiKey, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for User API Keys.
func (r *UserAPIKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Cannot update User API Key",
		"User API Keys need to be recreated to be updated",
	)
}

// Delete is not supported for User API Keys.
func (r *UserAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserAPIKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userClient, err := r.client.Users()
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("User", err))

		return
	}

	err = userClient.DeleteAPIKey(ctx, state.UserID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User API Key", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
