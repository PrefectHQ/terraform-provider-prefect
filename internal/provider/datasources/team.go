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

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &TeamDataSource{}
var _ datasource.DataSourceWithConfigure = &TeamDataSource{}

type TeamDataSource struct {
	client api.PrefectClient
}

type TeamDataSourceModel struct {
	BaseModel

	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`

	AccountID customtypes.UUIDValue `tfsdk:"account_id"`
}

// NewTeamDataSource returns a new TeamDataSource.
//
//nolint:ireturn // required by Terraform API
func NewTeamDataSource() datasource.DataSource {
	return &TeamDataSource{}
}

// Metadata returns the data source type name.
func (d *TeamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Shared set of schema attributes between team (singular)
// and teams (plural) datasources. Any team (singular)
// specific attributes will be added to a deep copy in the Schema method.
var teamAttributesBase = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Team ID (UUID)",
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the team creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the team was last updated in RFC 3339 format",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
		Description: "Name of Team",
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of team",
	},
}

// Schema defines the schema for the data source.
func (d *TeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Create a copy of the base attributes
	// and add the account ID overrides here
	// as they are not needed in the teams (plural) list
	teamAttributes := make(map[string]schema.Attribute)
	for k, v := range teamAttributesBase {
		teamAttributes[k] = v
	}
	teamAttributes["account_id"] = schema.StringAttribute{
		CustomType:  customtypes.UUIDType{},
		Description: "Account ID (UUID), defaults to the account set in the provider",
		Optional:    true,
	}

	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans(`
Get information about an existing Team by their name.
<br>
Use this data source to obtain team IDs to manage Workspace Access.
<br>
For more information, see [manage teams](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams).
`,
			helpers.PlanEnterprise,
		),
		Attributes: teamAttributes,
	}
}

// Configure adds the provider-configured client to the data source.
func (d *TeamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *TeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config TeamDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Teams(config.AccountID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Teams", err))

		return
	}

	// Fetch an existing Team by name
	// Here, we'd expect only 1 Team (or none) to be returned
	// as we are querying a single Team name, not a list of names
	// teams, err := client.List(ctx, []string{model.Name.ValueString()})
	teams, err := client.List(ctx, []string{config.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Team", "list", err))

		return
	}

	if len(teams) != 1 {
		resp.Diagnostics.AddError(
			"Could not find Team",
			fmt.Sprintf("Could not find Team with name %s", config.Name.ValueString()),
		)

		return
	}

	fetchedTeam := teams[0]

	config.ID = customtypes.NewUUIDValue(fetchedTeam.ID)
	config.Created = customtypes.NewTimestampPointerValue(fetchedTeam.Created)
	config.Updated = customtypes.NewTimestampPointerValue(fetchedTeam.Updated)
	config.Name = types.StringValue(fetchedTeam.Name)
	config.Description = types.StringPointerValue(fetchedTeam.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
