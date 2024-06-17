package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	ID          customtypes.UUIDValue      `tfsdk:"id"`
	Created     customtypes.TimestampValue `tfsdk:"created"`
	Updated     customtypes.TimestampValue `tfsdk:"updated"`
	AccountID   customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue      `tfsdk:"workspace_id"`

	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Tags  types.List   `tfsdk:"tags"`
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
	"value": schema.StringAttribute{
		Computed:    true,
		Description: "Value of the variable",
	},
	"tags": schema.ListAttribute{
		Computed:    true,
		Description: "Tags associated with the variable",
		ElementType: types.StringType,
	},
}

// Schema defines the schema for the data source.
func (d *VariableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Variable by name or ID.
<br>
Use this data source to obtain Variable-specific attributes, such as the value.
`,
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
		resp.Diagnostics.AddError(
			"Error creating variable client",
			fmt.Sprintf("Could not create variable client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

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
	model.Value = types.StringValue(variable.Value)

	list, diags := types.ListValueFrom(ctx, types.StringType, variable.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Tags = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
