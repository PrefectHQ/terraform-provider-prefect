package datasources

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = datasource.DataSourceWithConfigure(&VariableDataSource{})

// VariableDataSource contains state for the data source.
type VariableDataSource struct {
	client api.VariablesClient
}

// VariableDataSourceModel defines the Terraform data source model.
type VariableDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`

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
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client, _ = client.Variables(uuid.Nil, uuid.Nil)
}

var variableAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		Description: "Variable UUID",
		Optional:    true,
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
		Description: "Data Source representing a Prefect variable",
		Attributes:  variableAttributes,
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

		variable, err = d.client.Get(ctx, variableID)
	} else if !model.Name.IsNull() {
		variable, err = d.client.GetByName(ctx, model.Name.ValueString())
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
