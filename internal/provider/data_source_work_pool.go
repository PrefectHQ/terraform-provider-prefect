package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = datasource.DataSourceWithConfigure(&WorkPoolDataSource{})

type WorkPoolDataSource struct {
	client api.WorkPoolsClient
}

type WorkPoolSourceModel struct {
	Name    types.String `tfsdk:"name"`
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`
}

//nolint:ireturn
func NewWorkPoolDataSource() datasource.DataSource {
	return &WorkPoolDataSource{}
}

func (d *WorkPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client.WorkPools()
}

func (d *WorkPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pool"
}

func (d *WorkPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"updated": schema.StringAttribute{
				Computed: true,
			},
			// "description": schema.StringAttribute{
			// 	Computed: true,
			// },
			// "type": schema.StringAttribute{
			// 	Computed: true,
			// },
			// "base_job_template": schema.MapNestedAttribute{
			// 	Computed: true,
			// },
			// "paused": schema.BoolAttribute{
			// 	Computed: true,
			// },
			// "concurrency_limit": schema.Int64Attribute{
			// 	Computed: true,
			// },
			// "default_queue_id": schema.StringAttribute{
			// 	Computed: true,
			// },
		},
	}
}

func (d *WorkPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WorkPoolSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := d.client.Get(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("error", err.Error())
	}

	data.ID = basetypes.NewStringValue(pool.ID.String())

	if pool.Created == nil {
		data.Created = basetypes.NewStringNull()
	} else {
		data.Created = basetypes.NewStringValue(pool.Created.Format(time.RFC3339))
	}

	if pool.Updated == nil {
		data.Updated = basetypes.NewStringNull()
	} else {
		data.Updated = basetypes.NewStringValue(pool.Updated.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
