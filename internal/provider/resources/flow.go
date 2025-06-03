package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var (
	_ = resource.ResourceWithConfigure(&FlowResource{})
	_ = resource.ResourceWithImportState(&FlowResource{})
)

// FlowResource contains state for the resource.
type FlowResource struct {
	client api.PrefectClient
}

// FlowResourceModel defines the Terraform resource model.
type FlowResourceModel struct {
	BaseModel

	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`
	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`

	Name types.String `tfsdk:"name"`
	Tags types.List   `tfsdk:"tags"`
}

// NewFlowResource returns a new FlowResource.
//
//nolint:ireturn // required by Terraform API
func NewFlowResource() resource.Resource {
	return &FlowResource{}
}

// Metadata returns the resource type name.
func (r *FlowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flow"
}

// Configure initializes runtime state for the resource.
func (r *FlowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider client type",
			fmt.Sprintf("Expected api.PrefectClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Schema defines the schema for the resource.
func (r *FlowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		Description: helpers.DescriptionWithPlans("The resource `flow` represents a Prefect Flow. "+
			"Flows are the most central Prefect object. "+
			"A flow is a container for workflow logic as-code and allows users to configure how their workflows behave. "+
			"Flows are defined as Python functions, and any Python function is eligible to be a flow. "+
			"For more information, see [write and run flows](https://docs.prefect.io/v3/develop/write-flows).",
			helpers.AllPlans...,
		),
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Flow ID (UUID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID)",
			},
			"name": schema.StringAttribute{
				Description: "Name of the flow",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"tags": schema.ListAttribute{
				Description: "Tags associated with the flow",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(defaultEmptyTagList),
			},
		},
	}
}

// copyFlowToModel copies an api.Flow to a FlowResourceModel.
func copyFlowToModel(ctx context.Context, flow *api.Flow, model *FlowResourceModel) diag.Diagnostics {
	model.ID = customtypes.NewUUIDValue(flow.ID)
	model.Created = customtypes.NewTimestampPointerValue(flow.Created)
	model.Updated = customtypes.NewTimestampPointerValue(flow.Updated)
	model.Name = types.StringValue(flow.Name)

	tags, diags := types.ListValueFrom(ctx, types.StringType, flow.Tags)
	if diags.HasError() {
		return diags
	}
	model.Tags = tags

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *FlowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FlowResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Flows(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Flow", err))

		return
	}

	flow, err := client.Create(ctx, api.FlowCreate{
		Name: plan.Name.ValueString(),
		Tags: tags,
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Flow", "create", err))

		return
	}

	resp.Diagnostics.Append(copyFlowToModel(ctx, flow, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *FlowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model FlowResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Flows(model.AccountID.ValueUUID(), model.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Flow", err))

		return
	}

	// A flow can be imported + read by specifying the workspace_id and the flow_id.
	// if the workspace_id is omitted, then the default workspace_id is used.
	var flow *api.Flow
	if !model.ID.IsNull() {
		var flowID uuid.UUID
		flowID, err = uuid.Parse(model.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Error parsing Flow ID",
				fmt.Sprintf("Could not parse Flow ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}

		flow, err = client.Get(ctx, flowID)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing flow state",
			fmt.Sprintf("Could not read Flow, unexpected error: %s", err.Error()),
		)

		return
	}

	resp.Diagnostics.Append(copyFlowToModel(ctx, flow, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *FlowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FlowResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Flows(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Flow", err))

		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flowID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Flow ID",
			fmt.Sprintf("Could not parse flow ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Update(ctx, flowID, api.FlowUpdate{
		Tags: tags,
	})
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Flow", "update", err))

		return
	}

	flow, err := client.Get(ctx, flowID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Flow", "update", err))

		return
	}

	resp.Diagnostics.Append(copyFlowToModel(ctx, flow, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *FlowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FlowResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.Flows(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("Flow", err))

		return
	}

	flowID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Flow ID",
			fmt.Sprintf("Could not parse flow ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	err = client.Delete(ctx, flowID)
	if err != nil {
		resp.Diagnostics.Append(helpers.ResourceClientErrorDiagnostic("Flow", "delete", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *FlowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	helpers.ImportStateByID(ctx, req, resp)
}
