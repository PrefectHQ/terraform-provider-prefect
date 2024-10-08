package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

type WorkerMetadataDataSource struct {
	client api.PrefectClient
}

type WorkerMetadataDataSourceModel struct {
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	BaseJobConfigs types.Object `tfsdk:"base_job_configs"`
}

// NewWorkerMetadataDataSource returns a new WorkerMetadataDataSource.
//
//nolint:ireturn // required by Terraform API
func NewWorkerMetadataDataSource() datasource.DataSource {
	return &WorkerMetadataDataSource{}
}

// Metadata returns the data source type name.
func (d *WorkerMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_metadata"
}

// Configure initializes runtime state for the data source.
func (d *WorkerMetadataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *WorkerMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get metadata information about the common Worker types, such as Kubernetes, ECS, etc.
<br>
Use this data source to get the default base job configurations for those common Worker types.
`,
		Attributes: map[string]schema.Attribute{
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
			"base_job_configs": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "A map of default base job configurations (JSON) for each of the primary worker types",
				Attributes: map[string]schema.Attribute{
					"kubernetes": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Kubernetes workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"ecs": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for ECS workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"azure_container_instances": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Azure Container Instances workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"docker": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Docker workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"cloud_run": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Cloud Run workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"cloud_run_v2": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Cloud Run V2 workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"vertex_ai": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Vertex AI workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"prefect_agent": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Prefect Agent workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"process": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Process workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"azure_container_instances_push": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Azure Container Instances Push workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"cloud_run_push": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Cloud Run Push workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"cloud_run_v2_push": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for Cloud Run V2 Push workers",
						CustomType:  jsontypes.NormalizedType{},
					},
					"ecs_push": schema.StringAttribute{
						Computed:    true,
						Description: "Default base job configuration for ECS Push workers",
						CustomType:  jsontypes.NormalizedType{},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *WorkerMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WorkerMetadataDataSourceModel

	// Populate the model from data source configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Collections(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Collections", err))

		return
	}

	workerTypeByPackage, err := client.GetWorkerMetadataViews(ctx)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Worker Metadata", "get", err))

		return
	}

	// Flatten + remap the response payload so that the result is
	// a map of worker types to their default base job configurations.
	remap := make(map[string]json.RawMessage)
	for _, metadataByWorkerType := range workerTypeByPackage {
		for workerType, metadata := range metadataByWorkerType {
			remap[workerType] = metadata.DefaultBaseJobConfiguration
		}
	}

	// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/object#setting-values
	attributeTypes := map[string]attr.Type{
		"kubernetes":                     jsontypes.NormalizedType{},
		"ecs":                            jsontypes.NormalizedType{},
		"azure_container_instances":      jsontypes.NormalizedType{},
		"docker":                         jsontypes.NormalizedType{},
		"cloud_run":                      jsontypes.NormalizedType{},
		"cloud_run_v2":                   jsontypes.NormalizedType{},
		"vertex_ai":                      jsontypes.NormalizedType{},
		"prefect_agent":                  jsontypes.NormalizedType{},
		"process":                        jsontypes.NormalizedType{},
		"azure_container_instances_push": jsontypes.NormalizedType{},
		"cloud_run_push":                 jsontypes.NormalizedType{},
		"cloud_run_v2_push":              jsontypes.NormalizedType{},
		"ecs_push":                       jsontypes.NormalizedType{},
	}
	attributeValues := map[string]attr.Value{
		"kubernetes":                     jsontypes.NewNormalizedValue(string(remap["kubernetes"])),
		"ecs":                            jsontypes.NewNormalizedValue(string(remap["ecs"])),
		"azure_container_instances":      jsontypes.NewNormalizedValue(string(remap["azure-container-instance"])),
		"docker":                         jsontypes.NewNormalizedValue(string(remap["docker"])),
		"cloud_run":                      jsontypes.NewNormalizedValue(string(remap["cloud-run"])),
		"cloud_run_v2":                   jsontypes.NewNormalizedValue(string(remap["cloud-run-v2"])),
		"vertex_ai":                      jsontypes.NewNormalizedValue(string(remap["vertex-ai"])),
		"prefect_agent":                  jsontypes.NewNormalizedValue(string(remap["prefect-agent"])),
		"process":                        jsontypes.NewNormalizedValue(string(remap["process"])),
		"azure_container_instances_push": jsontypes.NewNormalizedValue(string(remap["azure-container-instance:push"])),
		"cloud_run_push":                 jsontypes.NewNormalizedValue(string(remap["cloud-run:push"])),
		"cloud_run_v2_push":              jsontypes.NewNormalizedValue(string(remap["cloud-run-v2:push"])),
		"ecs_push":                       jsontypes.NewNormalizedValue(string(remap["ecs:push"])),
	}

	obj, diag := types.ObjectValue(attributeTypes, attributeValues)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.BaseJobConfigs = obj
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
