package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var (
	_ = resource.ResourceWithConfigure(&VariableResource{})
	_ = resource.ResourceWithImportState(&VariableResource{})
)

// VariableResource contains state for the resource.
type VariableResource struct {
	client api.VariablesClient
}

// VariableResourceModel defines the Terraform resource model.
type VariableResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`

	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Tags  types.List   `tfsdk:"tags"`
}

// NewVariableResource returns a new VariableResource.
//
//nolint:ireturn // required by Terraform API
func NewVariableResource() resource.Resource {
	return &VariableResource{}
}

// Metadata returns the resource type name.
func (r *VariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

// Configure initializes runtime state for the resource.
func (r *VariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client.Variables()
}

// Schema defines the schema for the resource.
func (r *VariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Variable UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of the variable creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time that the variable was last updated in RFC 3339 format",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the variable",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value of the variable",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Description: "Tags associated with the variable",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *VariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model VariableResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variable, err := r.client.Create(ctx, api.VariableCreate{
		Name:  model.Name.ValueString(),
		Value: model.Value.ValueString(),
		Tags:  tags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable",
			fmt.Sprintf("Could not create variable, unexpected error: %s", err),
		)

		return
	}

	model.ID = types.StringValue(variable.ID.String())
	model.Name = types.StringValue(variable.Name)

	if variable.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(variable.Created.Format(time.RFC3339))
	}

	if variable.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(variable.Updated.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *VariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model VariableResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Always prefer to refresh state using the ID, if it is set.
	//
	// If we are importing by name, then we will need to load once using the name.
	var variable *api.Variable
	var err error
	if !model.ID.IsNull() {
		var variableID uuid.UUID
		variableID, err = uuid.Parse(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error parsing Variable ID",
				fmt.Sprintf("Could not parse variable ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}

		variable, err = r.client.Get(ctx, variableID)
	} else if !model.Name.IsNull() {
		variable, err = r.client.GetByName(ctx, model.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing variable state",
			fmt.Sprintf("Could not read variable, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = types.StringValue(variable.ID.String())
	model.Name = types.StringValue(variable.Name)
	model.Value = types.StringValue(variable.Value)

	if variable.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(variable.Created.Format(time.RFC3339))
	}

	if variable.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(variable.Updated.Format(time.RFC3339))
	}

	var diags diag.Diagnostics
	model.Tags, diags = types.ListValueFrom(ctx, types.StringType, variable.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model VariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Variable ID",
			fmt.Sprintf("Could not parse variable ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	var tags []string
	resp.Diagnostics.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = r.client.Update(ctx, variableID, api.VariableUpdate{
		Name:  model.Name.ValueString(),
		Value: model.Value.ValueString(),
		Tags:  tags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating variable",
			fmt.Sprintf("Could not update variable, unexpected error: %s", err),
		)

		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *VariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model VariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Variable ID",
			fmt.Sprintf("Could not parse variable ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = r.client.Delete(ctx, variableID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting variable",
			fmt.Sprintf("Could not delete variable, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *VariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.HasPrefix(req.ID, "name/") {
		name := strings.TrimPrefix(req.ID, "name/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}
