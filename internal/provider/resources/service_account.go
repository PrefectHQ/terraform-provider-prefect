package resources

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// ServiceAccountResource contains state for the resource.
type ServiceAccountResource struct {
	client api.PrefectClient
}

// ServiceAccountResourceModel defines the Terraform resource model.
type ServiceAccountResourceModel struct {
	ID   types.String               `tfsdk:"id"`
	Name types.String               `tfsdk:"name"`
	// Add more properties of Service Account here...
}

// NewServiceAccountResource returns a new ServiceAccountResource.
func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

// Metadata returns the resource type name.
func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

// Configure initializes runtime state for the resource.
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

// Schema defines the schema for the resource.
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
				Required:    true,
				Description: "Name of the service account",
			},
			// Add more properties of Service Account here...
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Service Account API to create a service account...
	serviceAccount, err := r.client.CreateServiceAccount(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service account",
			fmt.Sprintf("Could not create service account, unexpected error: %s", err),
		)

		return
	}

	// Set the ID of the created service account
	model.ID = types.StringValue(serviceAccount.ID)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the service account from API...
	serviceAccount, err := r.client.GetServiceAccount(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing service account state",
			fmt.Sprintf("Could not read service account, unexpected error: %s", err),
		)

		return
	}

	// Update the model with the latest data
	model.Name = types.StringValue(serviceAccount.Name)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}


// Update updates the resource with new data.
func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from the new plan and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the service account in API...
	_, err := r.client.UpdateServiceAccount(ctx, model.ID.ValueString(), model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service account",
			fmt.Sprintf("Could not update service account, unexpected error: %s", err),
		)

		return
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete removes the resource.
func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ServiceAccountResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the service account in API...
	err := r.client.DeleteServiceAccount(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting service account",
			fmt.Sprintf("Could not delete service account, unexpected error: %s", err),
		)

		return
	}
}

