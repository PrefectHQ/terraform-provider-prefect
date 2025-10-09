package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = datasource.DataSourceWithConfigure(&VariableDataSource{})

// VariableDataSource contains state for the data source.
type VariableDataSource struct {
	client api.PrefectClient
}

// VariableDataSourceModel defines the Terraform data source model.
type VariableDataSourceModel struct {
	BaseModel

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	Name  types.String  `tfsdk:"name"`
	Value types.Dynamic `tfsdk:"value"`
	Tags  types.Set     `tfsdk:"tags"`
}

// NewVariableDataSource returns a new VariableDataSource.
//
//nolint:ireturn // required by Terraform API
func NewVariableDataSource() datasource.DataSource {
	return &VariableDataSource{}
}

// Metadata returns the data source type name.
func (d *VariableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

// Configure initializes runtime state for the data source.
func (d *VariableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.Append(helpers.ConfigureTypeErrorDiagnostic("data source", req.ProviderData))

		return
	}

	d.client = client
}

var variableAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Variable ID (UUID)",
		Optional:    true,
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
		Computed:    true,
		Description: "Name of the variable",
		Optional:    true,
	},
	"value": schema.DynamicAttribute{
		Computed:    true,
		Description: "Value of the variable, supported Terraform value types: string, number, bool, tuple, object",
	},
	"tags": schema.SetAttribute{
		Computed:    true,
		Description: "Tags associated with the variable",
		ElementType: types.StringType,
	},
}

// Schema defines the schema for the data source.
func (d *VariableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Variable by name or ID.
<br>
Use this data source to obtain Variable-specific attributes, such as the value.
<br>
For more information, see [get and set variables](https://docs.prefect.io/v3/develop/variables).
`,
			helpers.AllPlans...,
		),
		Attributes: variableAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *VariableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model VariableDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !model.ID.IsNull() && !model.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting variable lookup keys",
			"Variables can be identified by their unique name or ID, but not both.",
		)

		return
	}

	client, err := d.client.Variables(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Variable", err))

		return
	}

	var variable *api.Variable

	switch {
	case !model.ID.IsNull():
		variable, err = client.Get(ctx, model.ID.ValueUUID())
	case !model.Name.IsNull():
		variable, err = client.GetByName(ctx, model.Name.ValueString())
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

	model.ID = customtypes.NewUUIDValue(variable.ID)
	model.Created = customtypes.NewTimestampPointerValue(variable.Created)
	model.Updated = customtypes.NewTimestampPointerValue(variable.Updated)

	model.Name = types.StringValue(variable.Name)

	value, diags := getDynamicValue(model, variable)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Value = value

	set, diags := types.SetValueFrom(ctx, types.StringType, variable.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Tags = set

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// getDynamicValue converts the 'value' attribute from a native Go type to a DynamicValue.
func getDynamicValue(model VariableDataSourceModel, variable *api.Variable) (basetypes.DynamicValue, diag.Diagnostics) {
	var result basetypes.DynamicValue
	var diags diag.Diagnostics

	switch value := variable.Value.(type) {
	case string:
		result = types.DynamicValue(types.StringValue(value))

	case float64:
		result = types.DynamicValue(types.Float64Value(value))

	case bool:
		result = types.DynamicValue(types.BoolValue(value))

	case map[string]interface{}:
		byteSlice, err := json.Marshal(value)
		if err != nil {
			diags = append(diags, helpers.SerializeDataErrorDiagnostic("data", "Variable Value", err))
		}

		model.Value = types.DynamicValue(jsontypes.NewNormalizedValue(string(byteSlice)))

	case []interface{}:
		tupleTypes := make([]attr.Type, len(value))
		tupleValues := make([]attr.Value, len(value))

		for i, v := range value {
			// For now, we only support string values in tuples.
			// This can be expanded in the future when we're ready to type check
			// inside of a type check :).
			tupleTypes[i] = types.StringType

			val, ok := v.(string)
			if !ok {
				diags.Append(helpers.SerializeDataErrorDiagnostic("data", "Variable Value", fmt.Errorf("unable to convert variable value to string")))
			}

			tupleValues[i] = types.StringValue(val)
		}

		tupleValue, diags := types.TupleValue(tupleTypes, tupleValues)
		if diags.HasError() {
			diags.Append(diags...)

			return result, diags
		}

		result = types.DynamicValue(tupleValue)

	default:
		diags.Append(helpers.ResourceClientErrorDiagnostic("Variable", "type", fmt.Errorf("unsupported type: %T", value)))
	}

	return result, diags
}
