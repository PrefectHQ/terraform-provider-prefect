package datasources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &deploymentDataSource{}
	_ datasource.DataSourceWithConfigure = &deploymentDataSource{}
)

// deploymentDataSource is the data source implementation.
type deploymentDataSource struct {
	client api.PrefectClient
}

// DeploymentDataSourceModel defines the Terraform data source model.
type DeploymentDataSourceModel struct {
	// The model requires the same fields, so reuse the fields defined in the resource model.
	resources.DeploymentResourceModel
}

// NewDeploymentDataSource is a helper function to simplify the provider implementation.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentDataSource() datasource.DataSource {
	return &deploymentDataSource{}
}

// Metadata returns the data source type name.
func (d *deploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

// Schema defines the scema for the data source.
func (d *deploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
Get information about an existing Deployment by either:
- deployment ID, or
- deployment name
The Deployment ID takes precedence over deployment name.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment ID (UUID)",
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
				Description: "Workspace ID (UUID) to associate deployment to",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the deployment",
				Optional:    true,
			},
			"flow_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Flow ID (UUID) to associate deployment to",
				Optional:    true,
			},
			"paused": schema.BoolAttribute{
				Description: "Whether or not the deployment is paused.",
				Optional:    true,
				Computed:    true,
			},
			"enforce_parameter_schema": schema.BoolAttribute{
				Description: "Whether or not the deployment should enforce the parameter schema.",
				Optional:    true,
				Computed:    true,
			},
			"storage_document_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "ID of the associated storage document (UUID)",
			},
			"manifest_path": schema.StringAttribute{
				Description: "The path to the flow's manifest file, relative to the chosen storage.",
				Optional:    true,
				Computed:    true,
			},
			"job_variables": schema.StringAttribute{
				Description: "Overrides for the flow's infrastructure configuration.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"work_queue_name": schema.StringAttribute{
				Description: "The work queue for the deployment. If no work queue is set, work will not be scheduled.",
				Optional:    true,
				Computed:    true,
			},
			"work_pool_name": schema.StringAttribute{
				Description: "The name of the deployment's work pool.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the deployment.",
				Optional:    true,
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path to the working directory for the workflow, relative to remote storage or an absolute path.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "An optional version for the deployment.",
				Optional:    true,
				Computed:    true,
			},
			"entrypoint": schema.StringAttribute{
				Description: "The path to the entrypoint for the workflow, relative to the path.",
				Optional:    true,
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the deployment",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "Parameters for flow runs scheduled by the deployment.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"parameter_openapi_schema": schema.StringAttribute{
				Description: "The parameter schema of the flow, including defaults.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"concurrency_limit": schema.Int64Attribute{
				Description: "The deployment's concurrency limit.",
				Optional:    true,
				Computed:    true,
			},
			"concurrency_options": schema.SingleNestedAttribute{
				Description: "Concurrency options for the deployment.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"collision_strategy": schema.StringAttribute{
						Description: "Enumeration of concurrency collision strategies.",
						Required:    true,
					},
				},
			},
			// Pull steps are polymorphic and can have different schemas based on the pull step type.
			// In the resource schema, we only make `type` required. The other attributes are needed
			// based on the pull step type, which we'll validate in the resource layer.
			"pull_steps": schema.ListNestedAttribute{
				Description: "Pull steps to prepare flows for a deployment run.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of pull step",
							Required:    true,
						},
						"credentials": schema.StringAttribute{
							Description: "Credentials to use for the pull step. Refer to a {GitHub,GitLab,BitBucket} credentials block.",
							Optional:    true,
						},
						"requires": schema.StringAttribute{
							Description: "A list of Python package dependencies.",
							Optional:    true,
						},
						"directory": schema.StringAttribute{
							Description: "(For type 'set_working_directory') The directory to set as the working directory.",
							Optional:    true,
						},
						"repository": schema.StringAttribute{
							Description: "(For type 'git_clone') The URL of the repository to clone.",
							Optional:    true,
						},
						"branch": schema.StringAttribute{
							Description: "(For type 'git_clone') The branch to clone. If not provided, the default branch is used.",
							Optional:    true,
						},
						"access_token": schema.StringAttribute{
							Description: "(For type 'git_clone') Access token for the repository. Refer to a credentials block for security purposes. Used in leiu of 'credentials'.",
							Optional:    true,
						},
						"bucket": schema.StringAttribute{
							Description: "(For type 'pull_from_*') The name of the bucket where files are stored.",
							Optional:    true,
						},
						"folder": schema.StringAttribute{
							Description: "(For type 'pull_from_*') The folder in the bucket where files are stored.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *deploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model DeploymentDataSourceModel

	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.client.Deployments(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Deployment", err))
	}

	// A deployment can be imported + read by either ID or Handle
	// If both are set, we prefer the ID
	var deployment *api.Deployment
	var operation string
	if !model.ID.IsNull() {
		var deploymentID uuid.UUID
		deploymentID, err = uuid.Parse(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error parsing Deployment ID",
				fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}

		operation = "get"
		deployment, err = client.Get(ctx, deploymentID)
	} else if !model.Name.IsNull() {
		var deployments []*api.Deployment
		operation = "list"
		deployments, err = client.List(ctx, []string{model.Name.ValueString()})

		if len(deployments) != 1 {
			resp.Diagnostics.AddError(
				"Could not find Deployment",
				fmt.Sprintf("Could not find Deployment with name %s", model.Name.ValueString()),
			)

			return
		}

		deployment = deployments[0]
	}

	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Deployment", operation, err))

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parametersByteSlice, err := json.Marshal(deployment.Parameters)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("parameters", "Deployment parameters", err))
	}
	model.Parameters = jsontypes.NewNormalizedValue(string(parametersByteSlice))

	jobVariablesByteSlice, err := json.Marshal(deployment.JobVariables)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("job_variables", "Deployment job variables", err))
	}
	model.JobVariables = jsontypes.NewNormalizedValue(string(jobVariablesByteSlice))

	parameterOpenAPISchemaByteSlice, err := json.Marshal(deployment.ParameterOpenAPISchema)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("parameter_openapi_schema", "Deployment parameter OpenAPI schema", err))
	}
	model.ParameterOpenAPISchema = jsontypes.NewNormalizedValue(string(parameterOpenAPISchemaByteSlice))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure initializes runtime state for the data source.
func (d *deploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// copyDeploymentToModel leverages the function by the same name from the resources package to avoid repeating
// the logic. Because DeploymentResourceModel embeds the resource.DeploymentResourceModel type, we can cast
// it to the compatible type before calling the referenced function.
func copyDeploymentToModel(ctx context.Context, deployment *api.Deployment, model *DeploymentDataSourceModel) diag.Diagnostics {
	// We need to copy the DeploymentResourceModel fields to the
	// DeploymentDataSourceModel.  Unfortunately, we can't directly convert the
	// type because struct embedding does not automatically make the embedding
	// struct convertible to the embedded type.
	compatibleModel := &resources.DeploymentResourceModel{
		ID:      model.ID,
		Created: model.Created,
		Updated: model.Updated,

		AccountID:              model.AccountID,
		ConcurrencyLimit:       model.ConcurrencyLimit,
		ConcurrencyOptions:     model.ConcurrencyOptions,
		Description:            model.Description,
		EnforceParameterSchema: model.EnforceParameterSchema,
		Entrypoint:             model.Entrypoint,
		FlowID:                 model.FlowID,
		JobVariables:           model.JobVariables,
		ManifestPath:           model.ManifestPath,
		Name:                   model.Name,
		ParameterOpenAPISchema: model.ParameterOpenAPISchema,
		Parameters:             model.Parameters,
		Path:                   model.Path,
		Paused:                 model.Paused,
		PullSteps:              model.PullSteps,
		StorageDocumentID:      model.StorageDocumentID,
		Tags:                   model.Tags,
		Version:                model.Version,
		WorkPoolName:           model.WorkPoolName,
		WorkQueueName:          model.WorkQueueName,
		WorkspaceID:            model.WorkspaceID,
	}

	diags := resources.CopyDeploymentToModel(ctx, deployment, compatibleModel)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	// Pass the values back to the model for the data source.
	model.ID = compatibleModel.ID
	model.Created = compatibleModel.Created
	model.Updated = compatibleModel.Updated
	model.AccountID = compatibleModel.AccountID
	model.ConcurrencyLimit = compatibleModel.ConcurrencyLimit
	model.ConcurrencyOptions = compatibleModel.ConcurrencyOptions
	model.Description = compatibleModel.Description
	model.Entrypoint = compatibleModel.Entrypoint
	model.FlowID = compatibleModel.FlowID
	model.JobVariables = compatibleModel.JobVariables
	model.Name = compatibleModel.Name
	model.ParameterOpenAPISchema = compatibleModel.ParameterOpenAPISchema
	model.Parameters = compatibleModel.Parameters
	model.Path = compatibleModel.Path
	model.Paused = compatibleModel.Paused
	model.PullSteps = compatibleModel.PullSteps
	model.StorageDocumentID = compatibleModel.StorageDocumentID
	model.Version = compatibleModel.Version
	model.WorkPoolName = compatibleModel.WorkPoolName
	model.WorkQueueName = compatibleModel.WorkQueueName

	return nil
}