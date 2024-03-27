package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

var _ = datasource.DataSourceWithConfigure(&WorkQueueDataSource{})

// WorkQueueDataSource contains state for the data source.
type WorkQueueDataSource struct {
	client api.PrefectClient
}

// WorkQueueDataSourceModel defines the Terraform data source model.
type WorkQueueDataSourceModel struct {
	ID           customtypes.UUIDValue      `tfsdk:"id"`
	Created      customtypes.TimestampValue `tfsdk:"created"`
	Updated      customtypes.TimestampValue `tfsdk:"updated"`
	AccountID    customtypes.UUIDValue      `tfsdk:"account_id"`
	WorkspaceID  customtypes.UUIDValue      `tfsdk:"workspace_id"`
	WorkPoolName types.String               `tfsdk:"work_pool_name"`

	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	IsPaused         types.Bool   `tfsdk:"is_paused"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
	Priority         types.Int64  `tfsdk:"priority"`
	// not yet required by the api
	// WorkPoolID customtypes.UUIDValue  `tfsdk:"work_pool_id"`
	// LastPolled types.String    		  `tfsdk:"last_polled"`
	// Status     types.String           `tfsdk:"status"`
	// // WorkPool   *WorkPool              `tfsdk:"work_pool"`
}

// NewWorkQueueDataSource returns a new WorkQueueDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkQueueDataSource() datasource.DataSource {
	return &WorkQueueDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_queue"
}

// Configure initializes runtime state for the data source.
func (d *WorkQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

var workQueueAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.UUIDType{},
		Description: "Work queue ID (UUID)",
		Optional:    true,
	},
	"created": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time of the work queue creation in RFC 3339 format",
	},
	"updated": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "Date and time that the work queue was last updated in RFC 3339 format",
	},
	"work_pool_name": schema.StringAttribute{
		Computed:    true,
		CustomType:  customtypes.TimestampType{},
		Description: "The name of the work pool that the work queue belongs to",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Description: "Name of the work queue",
		Optional:    true,
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "Description of the work queue",
		Optional:    true,
	},
	"is_paused": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether this work queue is paused",
	},
	"concurrency_limit": schema.Int64Attribute{
		Computed:    true,
		Description: "The concurrency limit applied to this work queue",
		Optional:    true,
	},
	"priority": schema.Int64Attribute{
		Computed:    true,
		Description: "The work queue's priority",
		Optional:    true,
	},
}

// Schema defines the schema for the data source.
func (d *WorkQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Work Queue by name.
<br>
Use this data source to obtain Work Queue-specific attributes.
`,
		Attributes: workQueueAttributes,
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkQueueDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.WorkQueues(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID(), model.WorkPoolName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating variable client",
			fmt.Sprintf("Could not create variable client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	queue, err := client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work queue state",
			fmt.Sprintf("Could not read work queue, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = customtypes.NewUUIDValue(queue.ID)
	model.Created = customtypes.NewTimestampPointerValue(queue.Created)
	model.Updated = customtypes.NewTimestampPointerValue(queue.Updated)

	model.Description = types.StringPointerValue(queue.Description)
	model.IsPaused = types.BoolValue(queue.IsPaused)
	model.ConcurrencyLimit = types.Int64PointerValue(queue.ConcurrencyLimit)
	model.Priority = types.Int64PointerValue(queue.Priority)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
