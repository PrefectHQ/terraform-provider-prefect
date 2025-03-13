package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = resource.ResourceWithConfigure(&SLAResource{})

type SLAResource struct {
	client api.PrefectClient
}

type SLAResourceModel struct {
	ResourceID types.String `tfsdk:"resource_id"`
	SLAs       []SLAModel   `tfsdk:"slas"`

	AccountID   customtypes.UUIDValue `tfsdk:"account_id"`
	WorkspaceID customtypes.UUIDValue `tfsdk:"workspace_id"`
}

type SLAModel struct {
	Name          types.String         `tfsdk:"name"`
	Severity      types.String         `tfsdk:"severity"`
	Enabled       types.Bool           `tfsdk:"enabled"`
	OwnerResource types.String         `tfsdk:"owner_resource"`
	Duration      types.Int64          `tfsdk:"duration"`
	StaleAfter    types.Float64        `tfsdk:"stale_after"`
	Within        types.Float64        `tfsdk:"within"`
	ExpectedEvent types.String         `tfsdk:"expected_event"`
	ResourceMatch jsontypes.Normalized `tfsdk:"resource_match"`
}

// NewSLAResource returns a new SLAResource.
//
//nolint:ireturn // required by Terraform API
func NewSLAResource() resource.Resource {
	return &SLAResource{}
}

func (r *SLAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_sla"
}

func (r *SLAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.PrefectClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected api.PrefectClient, got: "+fmt.Sprintf("%T", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SLAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
The resource 'resource_sla' represents a Prefect Resource SLA.
<br>
For more information, see documentation on setting up [Service Level Agreements](https://docs.prefect.io/v3/automate/events/slas) on Prefect resources.
`,
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Description: "The ID of the Prefect resource to set the SLA for, in the format of `prefect.<resource_type>.<resource_id>`.",
				Required:    true,
			},
			"slas": schema.ListNestedAttribute{
				Description: "List of SLAs to set for the resource. Note that this is a declarative list, and any SLAs that are not defined in this list will be removed from the resource (if they existed previously). Existing SLAs will be updated to match the definitions in this list. See documentation on [Defining SLAs](https://docs.prefect.io/v3/automate/events/slas#defining-slas) for more information, as well as the [API specification](https://app.prefect.cloud/api/docs#tag/SLAs/operation/apply_slas_api_accounts__account_id__workspaces__workspace_id__slas_apply_resource_slas__resource_id__post) for the SLA payload structure.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the SLA",
						},
						"severity": schema.StringAttribute{
							Optional:    true,
							Description: "Severity level of the SLA. Can be one of `minor`, `low`, `moderate`, `high`, or `critical`. Defaults to `high`.",
							Validators: []validator.String{
								stringvalidator.OneOf("minor", "low", "moderate", "high", "critical"),
							},
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether the SLA is enabled",
						},
						"owner_resource": schema.StringAttribute{
							Optional:    true,
							Description: "Resource that owns this SLA",
						},
						"duration": schema.Int64Attribute{
							Optional:    true,
							Description: "(TimeToCompletion SLA) The maximum flow run duration in seconds allowed before the SLA is violated.",
						},
						"stale_after": schema.Float64Attribute{
							Optional:    true,
							Description: "(Frequency SLA) The amount of time after a flow run is considered stale.",
						},
						"within": schema.Float64Attribute{
							Optional:    true,
							Description: "(Freshness SLA or Lateness SLA) The amount of time after a flow run is considered stale or late.",
						},
						"expected_event": schema.StringAttribute{
							Optional:    true,
							Description: "(Freshness SLA) The event to expect for this SLA.",
						},
						"resource_match": schema.StringAttribute{
							Optional:    true,
							Description: "(Freshness SLA) The resource to match for this SLA. Use `jsonencode()`",
							CustomType:  jsontypes.NormalizedType{},
						},
					},
				},
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Account ID (UUID), defaults to the account set in the provider",
			},
			"workspace_id": schema.StringAttribute{
				Optional:    true,
				CustomType:  customtypes.UUIDType{},
				Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
			},
		},
	}
}

// convertSLAModelToSLAUpsertPayload converts a SLAModel to an api.SLAUpsert payload.
func convertSLAModelToSLAUpsertPayload(model SLAModel) (api.SLAUpsert, diag.Diagnostics) {
	sla := api.SLAUpsert{
		Name:          model.Name.ValueString(),
		Severity:      model.Severity.ValueStringPointer(),
		Enabled:       model.Enabled.ValueBoolPointer(),
		OwnerResource: model.OwnerResource.ValueStringPointer(),
		Duration:      model.Duration.ValueInt64Pointer(),
		StaleAfter:    model.StaleAfter.ValueFloat64Pointer(),
		Within:        model.Within.ValueFloat64Pointer(),
		ExpectedEvent: model.ExpectedEvent.ValueStringPointer(),
	}

	resourceMatch, diagnostics := helpers.UnmarshalOptional(model.ResourceMatch)
	if diagnostics.HasError() {
		return api.SLAUpsert{}, diagnostics
	}

	if len(resourceMatch) > 0 {
		sla.ResourceMatch = &resourceMatch
	}

	return sla, nil
}

func (r *SLAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SLAResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.SLAs(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("SLAs", err))

		return
	}

	// Convert the SLA models to api.SLAUpsert structs
	slasToApply := make([]api.SLAUpsert, 0, len(plan.SLAs))
	for i := range plan.SLAs {
		payload, diagnostics := convertSLAModelToSLAUpsertPayload(plan.SLAs[i])
		if diagnostics.HasError() {
			resp.Diagnostics.Append(diagnostics...)

			return
		}

		slasToApply = append(slasToApply, payload)
	}

	_, err = client.ApplyResourceSLAs(ctx, plan.ResourceID.ValueString(), slasToApply)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Applying SLAs",
			"Could not apply SLAs to resource: "+err.Error(),
		)

		return
	}

	// Save the plan
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SLAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// For now, we'll just use the state as-is, since the Terraform is modeled
	// after the API response, and the API response doesn't include the payload in the response.
	var state SLAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SLAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SLAResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.SLAs(plan.AccountID.ValueUUID(), plan.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("SLAs", err))

		return
	}

	// Convert the SLA models to api.SLAUpsert structs
	slasToApply := make([]api.SLAUpsert, 0, len(plan.SLAs))
	for i := range plan.SLAs {
		payload, diagnostics := convertSLAModelToSLAUpsertPayload(plan.SLAs[i])
		if diagnostics.HasError() {
			resp.Diagnostics.Append(diagnostics...)

			return
		}

		slasToApply = append(slasToApply, payload)
	}

	_, err = client.ApplyResourceSLAs(ctx, plan.ResourceID.ValueString(), slasToApply)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Applying SLAs",
			"Could not apply SLAs to resource: "+err.Error(),
		)

		return
	}

	// Save the plan
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SLAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SLAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// To delete all SLAs, we apply an empty list
	client, err := r.client.SLAs(state.AccountID.ValueUUID(), state.WorkspaceID.ValueUUID())
	if err != nil {
		resp.Diagnostics.Append(helpers.CreateClientErrorDiagnostic("SLAs", err))

		return
	}

	_, err = client.ApplyResourceSLAs(ctx, state.ResourceID.ValueString(), []api.SLAUpsert{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting SLAs",
			"Could not delete SLAs from resource: "+err.Error(),
		)

		return
	}
}
