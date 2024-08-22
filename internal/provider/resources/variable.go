package resources

import (
	"context"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

var (
	_ = resource.ResourceWithConfigure(&VariableResource{})
	_ = resource.ResourceWithImportState(&VariableResource{})
)

// VariableResource contains state for the resource.
type VariableResource struct {
	client api.PrefectClient
}

// VariableResourceModel defines the Terraform resource model.
type VariableResourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

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
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("resource", req.ProviderData))

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *VariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		Description: "The resource `variable` represents a Prefect Cloud Variable. " +
			"Variables enable you to store and reuse non-sensitive information in your flows. ",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				// We cannot use a CustomType due to a conflict with PlanModifiers; see
				// https://github.com/hashicorp/terraform-plugin-framework/issues/763
				// https://github.com/hashicorp/terraform-plugin-framework/issues/754
				Description: "Variable ID (UUID)",
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
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the variable",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Value of the variable",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the variable",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyTagList),
			},
		},
	}
}

// copyVariableToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyVariableToModel(ctx context.Context, variable *api.Variable, tfModel *VariableResourceModel) diag.Diagnostics {
	tfModel.ID = types.StringValue(variable.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(variable.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(variable.Updated)

	tfModel.Name = types.StringValue(variable.Name)
	tfModel.Value = types.StringValue(variable.Value)

	tags, diags := types.ListValueFrom(ctx, types.StringType, variable.Tags)
	if diags.HasError() {
		return diags
	}
	tfModel.Tags = tags

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *VariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VariableResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Variables(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	variable, err := retry.DoWithData(
		func() (*api.Variable, error) {
			return client.Create(ctx, api.VariableCreate{
				Name:  plan.Name.ValueString(),
				Value: plan.Value.ValueString(),
				Tags:  tags,
			})
		},
		utils.DefaultRetryOptions...,
	)

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Variable", "create", err))

		return
	}

	resp.Diagnostics.Append(copyVariableToModel(ctx, variable, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *VariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VariableResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Variables(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	// Always prefer to refresh state using the ID, if it is set.
	//
	// If we are importing by name, then we will need to load once using the name.
	var variable *api.Variable

	switch {
	case !state.ID.IsNull():
		var variableID uuid.UUID
		variableID, err = uuid.Parse(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Variable", err))

			return
		}
		variable, err = client.Get(ctx, variableID)
	case !state.Name.IsNull():
		variable, err = client.GetByName(ctx, state.Name.ValueString())
	default:
		resp.Diagnostics.AddError(
			"Both ID and Name are unset",
			"This is a bug in the Terraform provider. Please report it to the maintainers.",
		)

		return
	}

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Variable", "get", err))

		return
	}

	resp.Diagnostics.Append(copyVariableToModel(ctx, variable, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Variables(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variableID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Variable", err))

		return
	}

	err = client.Update(ctx, variableID, api.VariableUpdate{
		Name:  plan.Name.ValueString(),
		Value: plan.Value.ValueString(),
		Tags:  tags,
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Variable", "update", err))

		return
	}

	variable, err := client.Get(ctx, variableID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Variable", "get", err))

		return
	}

	resp.Diagnostics.Append(copyVariableToModel(ctx, variable, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *VariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Variables(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	variableID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Variable", err))

		return
	}

	err = client.Delete(ctx, variableID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Variable", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
// Valid import IDs:
// name/<variable_name>
// name/<variable_name>,<workspace_id>
// <variable_id>
// <variable_id>,<workspace_id>.
func (r *VariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")

	if len(parts) > 2 || len(parts) == 0 {
		resp.Diagnostics.AddError(
			"Error importing variable",
			"Import ID must be in the format of <variable identifier> OR <variable identifier>,<workspace_id>",
		)

		return
	}

	variableIdentifier := parts[0]

	if strings.HasPrefix(variableIdentifier, "name/") {
		name := strings.TrimPrefix(variableIdentifier, "name/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), variableIdentifier)...)
	}

	if len(parts) == 2 && parts[1] != "" {
		workspaceID, err := uuid.Parse(parts[1])
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Workspace", err))

			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID.String())...)
	}
}
