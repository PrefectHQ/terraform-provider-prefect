package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var (
	_ = resource.ResourceWithConfigure(&AutomationsResource{})
	_ = resource.ResourceWithImportState(&AutomationsResource{})
)

// AutomationsResource contains state for the resource.
type AutomationsResource struct {
	client api.PrefectClient
}

// NewAutomationsResource returns a new AutomationsResource.
//
//nolint:ireturn // required by Terraform API
func NewAutomationsResource() resource.Resource {
	return &AutomationsResource{}
}

// Metadata returns the resource type name.
func (r *AutomationsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation"
}

// Configure initializes runtime state for the resource.
func (r *AutomationsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected api.PrefectClient, got: "+fmt.Sprintf("%T", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *AutomationsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `automations` represents a Prefect Automation.",
		Version:     0,
		Attributes:  AutomationSchema(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *AutomationsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read refreshes the Terraform state with the latest data.
func (r *AutomationsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AutomationsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AutomationsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// ImportState imports the resource into Terraform state.
func (r *AutomationsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
