package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&DeploymentResource{})
	_ = resource.ResourceWithImportState(&DeploymentResource{})
)

// DeploymentResource contains state for the resource.
type DeploymentResource struct {
	client api.PrefectClient
}

// DeploymentResourceModel defines the Terraform resource model.
type DeploymentResourceModel struct {
	ID      types.String               `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`

	ConcurrencyLimit       types.Int64           `tfsdk:"concurrency_limit"`
	ConcurrencyOptions     types.Object          `tfsdk:"concurrency_options"`
	Description            types.String          `tfsdk:"description"`
	EnforceParameterSchema types.Bool            `tfsdk:"enforce_parameter_schema"`
	Entrypoint             types.String          `tfsdk:"entrypoint"`
	FlowID                 customtypes.UUIDValue `tfsdk:"flow_id"`
	JobVariables           jsontypes.Normalized  `tfsdk:"job_variables"`
	ManifestPath           types.String          `tfsdk:"manifest_path"`
	Name                   types.String          `tfsdk:"name"`
	ParameterOpenAPISchema jsontypes.Normalized  `tfsdk:"parameter_openapi_schema"`
	Parameters             jsontypes.Normalized  `tfsdk:"parameters"`
	Path                   types.String          `tfsdk:"path"`
	Paused                 types.Bool            `tfsdk:"paused"`
	StorageDocumentID      customtypes.UUIDValue `tfsdk:"storage_document_id"`
	Tags                   types.List            `tfsdk:"tags"`
	Version                types.String          `tfsdk:"version"`
	WorkPoolName           types.String          `tfsdk:"work_pool_name"`
	WorkQueueName          types.String          `tfsdk:"work_queue_name"`
}

// NewDeploymentResource returns a new DeploymentResource.
//
//nolint:ireturn // required by Terraform API
func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// Metadata returns the resource type name.
func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

// Configure initializes runtime state for the resource.
func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		// Description: "Resource representing a Prefect Workspace",
		Description: "Deployments are server-side representations of flows. " +
			"They store the crucial metadata needed for remote orchestration including when, where, and how a workflow should run. " +
			"Deployments elevate workflows from functions that you must call manually to API-managed entities that can be triggered remotely.",
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				// We cannot use a CustomType due to a conflict with PlanModifiers; see
				// https://github.com/hashicorp/terraform-plugin-framework/issues/763
				// https://github.com/hashicorp/terraform-plugin-framework/issues/754
				Description: "Workspace ID (UUID)",
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
				Description: "Workspace ID (UUID) to associate deployment to",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the workspace",
				Required:    true,
			},
			"flow_id": schema.StringAttribute{
				CustomType:  customtypes.UUIDType{},
				Description: "Flow ID (UUID) to associate deployment to",
				Required:    true,
			},
			"paused": schema.BoolAttribute{
				Description: "Whether or not the deployment is paused.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enforce_parameter_schema": schema.BoolAttribute{
				Description: "Whether or not the deployment should enforce the parameter schema.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"work_pool_name": schema.StringAttribute{
				Description: "The name of the deployment's work pool.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the deployment.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Description: "The path to the working directory for the workflow, relative to remote storage or an absolute path.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "An optional version for the deployment.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entrypoint": schema.StringAttribute{
				Description: "The path to the entrypoint for the workflow, relative to the path.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the deployment",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyTagList),
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
				// OpenAPI schema is also only set on create, and
				// we do not support modifying this value. Therefore, any changes
				// to this attribute will force a replacement.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"concurrency_limit": schema.Int64Attribute{
				Description: "The deployment's concurrency limit.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"concurrency_options": schema.SingleNestedAttribute{
				Description: "Concurrency options for the deployment.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"collision_strategy": schema.StringAttribute{
						Description: "Enumeration of concurrency collision strategies.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ENQUEUE", "CANCEL_NEW"),
						},
					},
				},
			},
		},
	}
}

// copyDeploymentToModel copies an api.Deployment to a DeploymentResourceModel.
func copyDeploymentToModel(ctx context.Context, deployment *api.Deployment, model *DeploymentResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(deployment.ID.String())
	model.Created = customtypes.NewTimestampPointerValue(deployment.Created)
	model.Updated = customtypes.NewTimestampPointerValue(deployment.Updated)

	model.Description = types.StringValue(deployment.Description)
	model.EnforceParameterSchema = types.BoolValue(deployment.EnforceParameterSchema)
	model.Entrypoint = types.StringValue(deployment.Entrypoint)
	model.FlowID = customtypes.NewUUIDValue(deployment.FlowID)
	model.ManifestPath = types.StringValue(deployment.ManifestPath)
	model.Name = types.StringValue(deployment.Name)
	model.Path = types.StringValue(deployment.Path)
	model.Paused = types.BoolValue(deployment.Paused)
	model.StorageDocumentID = customtypes.NewUUIDValue(deployment.StorageDocumentID)
	model.Version = types.StringValue(deployment.Version)
	model.WorkPoolName = types.StringValue(deployment.WorkPoolName)
	model.WorkQueueName = types.StringValue(deployment.WorkQueueName)

	tags, diags := types.ListValueFrom(ctx, types.StringType, deployment.Tags)
	if diags.HasError() {
		return diags
	}
	model.Tags = tags

	// The concurrency_limit field in the response payload is deprecated, and will always be 0
	// for compatibility. The true value has been moved under `global_concurrency_limit.limit`.
	if deployment.GlobalConcurrencyLimit != nil {
		model.ConcurrencyLimit = types.Int64Value(int64(deployment.GlobalConcurrencyLimit.Limit))
	}

	if deployment.ConcurrencyOptions != nil {
		concurrencyOptionsObject, diags := types.ObjectValue(
			map[string]attr.Type{
				"collision_strategy": types.StringType,
			},
			map[string]attr.Value{
				"collision_strategy": types.StringValue(deployment.ConcurrencyOptions.CollisionStrategy),
			},
		)
		if diags.HasError() {
			return diags
		}

		model.ConcurrencyOptions = concurrencyOptionsObject
	}

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeploymentResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameters, diags := helpers.SafeUnmarshal(plan.Parameters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobVariables, diags := helpers.SafeUnmarshal(plan.JobVariables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parameterOpenAPISchema, diags := helpers.SafeUnmarshal(plan.ParameterOpenAPISchema)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createPayload := api.DeploymentCreate{
		ConcurrencyLimit:       plan.ConcurrencyLimit.ValueInt64Pointer(),
		Description:            plan.Description.ValueString(),
		EnforceParameterSchema: plan.EnforceParameterSchema.ValueBool(),
		Entrypoint:             plan.Entrypoint.ValueString(),
		FlowID:                 plan.FlowID.ValueUUID(),
		JobVariables:           jobVariables,
		ManifestPath:           plan.ManifestPath.ValueString(),
		Name:                   plan.Name.ValueString(),
		Parameters:             parameters,
		Path:                   plan.Path.ValueString(),
		Paused:                 plan.Paused.ValueBool(),
		StorageDocumentID:      plan.StorageDocumentID.ValueUUIDPointer(),
		Tags:                   tags,
		Version:                plan.Version.ValueString(),
		WorkPoolName:           plan.WorkPoolName.ValueString(),
		WorkQueueName:          plan.WorkQueueName.ValueString(),
		ParameterOpenAPISchema: parameterOpenAPISchema,
	}

	if !plan.ConcurrencyOptions.IsNull() {
		concurrencyOptions := api.ConcurrencyOptions{
			CollisionStrategy: getUnescapedValue(plan.ConcurrencyOptions, "collision_strategy"),
		}

		createPayload.ConcurrencyOptions = &concurrencyOptions
	}

	deployment, err := client.Create(ctx, createPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment",
			fmt.Sprintf("Could not create deployment, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model DeploymentResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	// A deployment can be imported + read by either ID or Handle
	// If both are set, we prefer the ID
	var deployment *api.Deployment
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

		deployment, err = client.Get(ctx, deploymentID)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing deployment state",
			fmt.Sprintf("Could not read Deployment, unexpected error: %s", err.Error()),
		)

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

// Update updates the resource and sets the updated Terraform state on success.
func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model DeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)

		return
	}

	deploymentID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Deployment ID",
			fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	var tags []string
	resp.Diagnostics.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var parameters map[string]interface{}
	if !model.Parameters.IsNull() {
		resp.Diagnostics.Append(model.Parameters.Unmarshal(&parameters)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var jobVariables map[string]interface{}
	if !model.JobVariables.IsNull() {
		resp.Diagnostics.Append(model.JobVariables.Unmarshal(&jobVariables)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := api.DeploymentUpdate{
		ConcurrencyLimit:       model.ConcurrencyLimit.ValueInt64Pointer(),
		Description:            model.Description.ValueString(),
		EnforceParameterSchema: model.EnforceParameterSchema.ValueBool(),
		Entrypoint:             model.Entrypoint.ValueString(),
		JobVariables:           jobVariables,
		ManifestPath:           model.ManifestPath.ValueString(),
		Parameters:             parameters,
		Path:                   model.Path.ValueString(),
		Paused:                 model.Paused.ValueBool(),
		StorageDocumentID:      model.StorageDocumentID.ValueUUIDPointer(),
		Tags:                   tags,
		Version:                model.Version.ValueString(),
		WorkPoolName:           model.WorkPoolName.ValueString(),
		WorkQueueName:          model.WorkQueueName.ValueString(),
	}

	if !model.ConcurrencyOptions.IsNull() {
		concurrencyOptions := api.ConcurrencyOptions{
			CollisionStrategy: getUnescapedValue(model.ConcurrencyOptions, "collision_strategy"),
		}

		payload.ConcurrencyOptions = &concurrencyOptions
	}

	err = client.Update(ctx, deploymentID, payload)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating deployment",
			fmt.Sprintf("Could not update deployment, unexpected error: %s", err),
		)

		return
	}

	deployment, err := client.Get(ctx, deploymentID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Deployment state",
			fmt.Sprintf("Could not read Deployment, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyDeploymentToModel(ctx, deployment, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parametersByteSlice, err := json.Marshal(deployment.Parameters)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("parameters", "Deployment parameters", err))

		return
	}
	model.Parameters = jsontypes.NewNormalizedValue(string(parametersByteSlice))

	jobVariablesByteSlice, err := json.Marshal(deployment.JobVariables)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("job_variables", "Deployment job variables", err))

		return
	}
	model.JobVariables = jsontypes.NewNormalizedValue(string(jobVariablesByteSlice))

	parameterOpenAPISchemaByteSlice, err := json.Marshal(deployment.ParameterOpenAPISchema)
	if err != nil {
		resp.Diagnostics.Append(helpers.SerializeDataErrorDiagnostic("parameter_openapi_schema", "Deployment parameter OpenAPI schema", err))

		return
	}
	model.ParameterOpenAPISchema = jsontypes.NewNormalizedValue(string(parameterOpenAPISchemaByteSlice))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Deployments(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deployment client",
			fmt.Sprintf("Could not create deployment client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	deploymentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Deployment ID",
			fmt.Sprintf("Could not parse deployment ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, deploymentID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Deployment",
			fmt.Sprintf("Could not delete Deployment, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// we'll allow input values in the form of:
	// - "id,workspace_id"
	// - "id"
	maxInputCount := 2
	inputParts := strings.Split(req.ID, ",")

	// eg. "foo,bar,baz"
	if len(inputParts) > maxInputCount {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	// eg. ",foo" or "foo,"
	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected non-empty import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	identifier := inputParts[0]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), identifier)...)

	if len(inputParts) == 2 && inputParts[1] != "" {
		workspaceID, err := uuid.Parse(inputParts[1])
		if err != nil {
			resp.Diagnostics.Append(helpers.ParseUUIDErrorDiagnostic("Deployment", err))

			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID.String())...)
	}
}

// getUnescapedValue returns the unescaped value of a key in an Object.
// For some reason, without this function we see the value in the HTTP payload
// has escaped quotes. For example: "\"ENQUEUE\""
// This leads to Pydantic validation errors so we need to make sure we've stripped
// out any quotes from the value.
//
// There is very likely a better way to do this, or a way to avoid this entirely.
func getUnescapedValue(obj types.Object, key string) string {
	attrs := obj.Attributes()
	var result string

	if val, ok := attrs[key]; ok {
		result = strings.Trim(val.String(), `"`)

		// This didn't seem impactful in testing, but let's keep it to be safe.
		result = strings.ReplaceAll(result, `\"`, `"`)

		return result
	}

	return "problemescaping"
}
