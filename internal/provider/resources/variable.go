package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
)

var (
	_ = resource.ResourceWithConfigure(&VariableResource{})
	_ = resource.ResourceWithImportState(&VariableResource{})
	_ = resource.ResourceWithUpgradeState(&VariableResource{})
)

// VariableResource contains state for the resource.
type VariableResource struct {
	client api.PrefectClient
}

// VariableResourceModel defines the Terraform resource model.
// NOTE: we version the VersionResourceModel here due to a schema
// update, and we want to be able to properly migrate existing
// prefect_variable resources in state from one schema version to the next.
// See UpgradeState for more details.
//
// V0: Value is types.String.
type VariableResourceModelV0 struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Tags  types.List   `tfsdk:"tags"`
}

// V1: Value is types.Dynamic.
type VariableResourceModelV1 struct {
	ID          types.String               `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name  types.String  `tfsdk:"name"`
	Value types.Dynamic `tfsdk:"value"`
	Tags  types.List    `tfsdk:"tags"`
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
		Version: 1,
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
			"value": schema.DynamicAttribute{
				Description: "Value of the variable, supported Terraform value types: string, number, bool, tuple, object",
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

// UpgradeState adds upgraders to the VariableResource.
// This is needed when a resource schema change is made (eg. an attribute type).
// The key/index in the return object is the source version (eg. 0 -> current).
// The target version is the one defined in Schema.Version above
// https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade
func (r *VariableResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from prior (0) => current (Schema.Version)
		0: {
			// PriorSchema allows the framework to populate the req.State argument
			// for easier data handling when migrating existing state resources
			// to a new schema.
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"created": schema.StringAttribute{
						Computed:   true,
						CustomType: customtypes.TimestampType{},
					},
					"updated": schema.StringAttribute{
						Computed:   true,
						CustomType: customtypes.TimestampType{},
					},
					"account_id": schema.StringAttribute{
						CustomType: customtypes.UUIDType{},
						Optional:   true,
					},
					"workspace_id": schema.StringAttribute{
						CustomType: customtypes.UUIDType{},
						Optional:   true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						Required: true,
					},
					"tags": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     listdefault.StaticValue(defaultEmptyTagList),
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData VariableResourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

				if resp.Diagnostics.HasError() {
					return
				}

				// In order to update a prefect_variable resource in state
				// that is tied to the old schema version, we need to copy
				// the existing state into the new schema version.
				upgradedStateData := VariableResourceModelV1{
					ID:          priorStateData.ID,
					Created:     priorStateData.Created,
					Updated:     priorStateData.Updated,
					AccountID:   priorStateData.AccountID,
					WorkspaceID: priorStateData.WorkspaceID,
					Name:        priorStateData.Name,
					Tags:        priorStateData.Tags,
				}

				// This is the main upgrade operation between v0 => v1.
				// Convert the "value" attribute's type from
				// StringAttribute (v0) to a DynamicValue (v1)
				// to prevent a Terraform error when deserializing the state
				// from the old schema to the new one.
				upgradedStateData.Value = types.DynamicValue(basetypes.NewStringValue(priorStateData.Value.ValueString()))

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}

// copyVariableToModel maps an API response to a model that is saved in Terraform state.
// A model can be a Terraform Plan, State, or Config object.
func copyVariableToModel(ctx context.Context, variable *api.Variable, tfModel *VariableResourceModelV1) diag.Diagnostics {
	tfModel.ID = types.StringValue(variable.ID.String())
	tfModel.Created = customtypes.NewTimestampPointerValue(variable.Created)
	tfModel.Updated = customtypes.NewTimestampPointerValue(variable.Updated)

	tfModel.Name = types.StringValue(variable.Name)

	tags, diags := types.ListValueFrom(ctx, types.StringType, variable.Tags)
	if diags.HasError() {
		return diags
	}
	tfModel.Tags = tags

	return nil
}

// getUnderlyingValue converts the 'value' attribute from a DynamicValue to
// a native Go type that can be sent to the Prefect API.
func getUnderlyingValue(plan VariableResourceModelV1) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var value interface{}

	switch underlyingValue := plan.Value.UnderlyingValue().(type) {
	case types.String:
		value = underlyingValue.ValueString()

	case types.Number:
		var err error
		value, err = strconv.ParseFloat(underlyingValue.String(), 64)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"unable to convert number to float64",
				fmt.Sprintf("number: %v, error: %v", value, err),
			))
		}

	case types.Bool:
		value = underlyingValue.ValueBool()

	case types.Tuple:
		result := make([]string, len(underlyingValue.Elements()))
		for i, e := range underlyingValue.Elements() {
			result[i] = e.String()
		}

		value = result

	case types.Object:
		result := map[string]interface{}{}
		if err := json.Unmarshal([]byte(underlyingValue.String()), &result); err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"unable to convert object to map[string]interface",
				fmt.Sprintf("object: %v, error: %v", value, err),
			))
		}

		value = result

	default:
		diags.Append(diag.NewErrorDiagnostic(
			"unexpected value type",
			fmt.Sprintf("type: %T", underlyingValue),
		))
	}

	return value, diags
}

// Create creates the resource and sets the initial Terraform state.
func (r *VariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VariableResourceModelV1

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value, diags := getUnderlyingValue(plan)
	if diags.HasError() {
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

	variable, err := client.Create(ctx, api.VariableCreate{
		Name:  plan.Name.ValueString(),
		Value: value,
		Tags:  tags,
	})

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
	var state VariableResourceModelV1

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
	var plan VariableResourceModelV1

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Variables(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	value, diags := getUnderlyingValue(plan)
	if diags.HasError() {
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
		Value: value,
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
	var state VariableResourceModelV1

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
