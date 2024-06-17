package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = datasource.DataSourceWithConfigure(&TeamsDataSource{})

// TeamsDataSource contains state for the data source.
type TeamsDataSource struct {
	client api.PrefectClient
}

// TeamsDataSourceModel defines the Terraform data source model.
type TeamsDataSourceModel struct {
	AccountID customtypes.UUIDValue `tfsdk:"account_id"`

	Teams types.List `tfsdk:"teams"`
}

// NewTeamsDataSource returns a new TeamsDataSource.
//
//nolint:ireturn // required by Terraform API
func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

// Metadata returns the data source type name.
func (d *TeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

// Configure initializes runtime state for the data source.
func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Schema defines the schema for the data source.
func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about multiple Teams.
<br>
Use this data source to search for multiple Teams. Defaults to fetching all Teams in the Account.
`,
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
				Optional:    true,
			},
			"teams": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Teams returned by the server",
				NestedObject: schema.NestedAttributeObject{
					Attributes: teamAttributesBase,
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model TeamsDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Teams(model.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable client",
			fmt.Sprintf("Could not create variable client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	// Fetch all existing teams
	var filter []string
	teams, err := client.List(ctx, filter)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Teams", "list", err))

		return
	}

	attributeTypes := map[string]attr.Type{
		"id":          customtypes.UUIDType{},
		"created":     customtypes.TimestampType{},
		"updated":     customtypes.TimestampType{},
		"name":        types.StringType,
		"description": types.StringType,
	}

	teamObjects := make([]attr.Value, 0, len(teams))
	for _, team := range teams {
		attributeValues := map[string]attr.Value{
			"id":          customtypes.NewUUIDValue(team.ID),
			"created":     customtypes.NewTimestampPointerValue(team.Created),
			"updated":     customtypes.NewTimestampPointerValue(team.Updated),
			"name":        types.StringValue(team.Name),
			"description": types.StringValue(team.Description),
		}

		teamObject, diag := types.ObjectValue(attributeTypes, attributeValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		teamObjects = append(teamObjects, teamObject)
	}

	list, diag := types.ListValue(types.ObjectType{AttrTypes: attributeTypes}, teamObjects)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Teams = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
