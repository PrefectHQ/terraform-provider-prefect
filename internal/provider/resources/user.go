package resources

import (
	"context"

	"github.com/google/uuid"
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
	_ = resource.ResourceWithConfigure(&UserResource{})
	_ = resource.ResourceWithImportState(&UserResource{})
)

// UserResource contains state for the user resource.
type UserResource struct {
	client api.PrefectClient
}

// UserResourceModel defines the Terraform resource model.
type UserResourceModel struct {
	BaseModel

	// Read-only fields
	ActorID customtypes.UUIDValue `tfsdk:"actor_id"`

	// Updateable fields
	Handle    types.String `tfsdk:"handle"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
}

// NewUserResource returns a new UserResource.
//
//nolint:ireturn // required by Terraform API
func NewUserResource() resource.Resource {
	return &UserResource{}
}

// Metadata returns the resource type name.
func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Configure initializes runtime state for the resource.
func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(
			"The resource `user` represents a Prefect User.",
			helpers.AllPlans...,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "User ID (UUID)",
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
				Description: "Timestamp of when the resource was updated (RFC3339)",
				Computed:    true,
				CustomType:  customtypes.TimestampType{},
			},
			"actor_id": schema.StringAttribute{
				Description: "Actor ID (UUID), used for granting access to resources like Teams",
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
			},
			"handle": schema.StringAttribute{
				Description: "A unique handle for the user, containing only lowercase letters, numbers, and dashes.",
				Computed:    true,
				Optional:    true,
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user",
				Computed:    true,
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user",
				Computed:    true,
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email of the user",
				Computed:    true,
				Optional:    true,
			},
		},
	}
}

// Create creates a new user.
func (r *UserResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Not implemented", "Creating a user is not yet implemented")
}

// Read reads a user.
func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() {
		resp.Diagnostics.AddError(
			"ID is unset",
			"Ensure the user has been imported into the state before attempting to read.",
		)

		return
	}

	client, err := r.client.Users()
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("User", err))

		return
	}

	user, err := client.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User", "read", err))

		return
	}

	copyUserToModel(user, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates a user.
func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Users()
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("User", err))

		return
	}

	err = client.Update(ctx, plan.ID.ValueString(), api.UserUpdate{
		Handle:    plan.Handle.ValueString(),
		FirstName: plan.FirstName.ValueString(),
		LastName:  plan.LastName.ValueString(),
		Email:     plan.Email.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User", "update", err))

		return
	}

	user, err := client.Read(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("User", "read", err))

		return
	}

	copyUserToModel(user, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes a user.
// Delete is not implemented in the provider because Create is also not implemented.
// While there is a valid Delete endpoint for Users, this is not something a practitioner
// would typically want the provider to perform. User deletion is a more involved process
// that includes removing API keys, personal accounts, workspaces, and other resources.
func (r *UserResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func copyUserToModel(user *api.User, state *UserResourceModel) {
	state.ActorID = customtypes.NewUUIDValue(uuid.MustParse(user.ActorID))
	state.Handle = types.StringValue(user.Handle)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.Email = types.StringValue(user.Email)
}

// ImportState imports a user.
func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
