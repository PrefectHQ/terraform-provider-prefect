package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var (
	_ = resource.ResourceWithConfigure(&WorkPoolResource{})
	_ = resource.ResourceWithImportState(&WorkPoolResource{})
)

// WorkPoolResource contains state for the resource.
type WorkPoolResource struct {
	client api.WorkPoolsClient
}

// WorkPoolResourceModel defines the Terraform resource model.
type WorkPoolResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`

	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Type             types.String `tfsdk:"type"`
	Paused           types.Bool   `tfsdk:"paused"`
	ConcurrencyLimit types.Int64  `tfsdk:"concurrency_limit"`
	DefaultQueueID   types.String `tfsdk:"default_queue_id"`
	BaseJobTemplate  types.String `tfsdk:"base_job_template"`
}

// New returns a new WorkPoolResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkPoolResource() resource.Resource {
	return &WorkPoolResource{}
}

// Metadata returns the resource type name.
func (r *WorkPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_work_pool"
}

// Configure initializes runtime state for the resource.
func (r *WorkPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client.WorkPools()
}

// Schema defines the schema for the resource.
func (r *WorkPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Work pool UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of the work pool creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time that the work pool was last updated in RFC 3339 format",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the work pool",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the work pool",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Default:     stringdefault.StaticString("prefect-agent"),
				Description: "Type of the work pool",
				Optional:    true,
			},
			"paused": schema.BoolAttribute{
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether this work pool is paused",
				Optional:    true,
			},
			"concurrency_limit": schema.Int64Attribute{
				Description: "The concurrency limit applied to this work pool",
				Optional:    true,
			},
			"default_queue_id": schema.StringAttribute{
				Computed:    true,
				Description: "The UUID of the default queue associated with this work pool",
				Optional:    true,
			},
			"base_job_template": schema.StringAttribute{
				Computed:    true,
				Description: "The base job template for the work pool, as a JSON string",
				Optional:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model WorkPoolResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate := map[string]interface{}{}
	if !model.BaseJobTemplate.IsNull() {
		reader := strings.NewReader(model.BaseJobTemplate.ValueString())
		decoder := json.NewDecoder(reader)
		err := decoder.Decode(&baseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to deserialize Base Job Template",
				fmt.Sprintf("Failed to deserialize Base Job Template as JSON object: %s", err),
			)

			return
		}
	}

	pool, err := r.client.Create(ctx, api.WorkPoolCreate{
		Name:             model.Name.ValueString(),
		Description:      model.Description.ValueStringPointer(),
		Type:             model.Type.ValueString(),
		BaseJobTemplate:  baseJobTemplate,
		IsPaused:         model.Paused.ValueBool(),
		ConcurrencyLimit: model.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating work pool",
			fmt.Sprintf("Could not create work pool, unexpected error: %s", err),
		)

		return
	}

	model.ID = types.StringValue(pool.ID.String())
	model.Name = types.StringValue(pool.Name)

	if pool.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(pool.Created.Format(time.RFC3339))
	}

	if pool.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(pool.Updated.Format(time.RFC3339))
	}

	model.Description = types.StringPointerValue(pool.Description)
	model.Type = types.StringValue(pool.Type)
	model.Paused = types.BoolValue(pool.IsPaused)
	model.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	model.DefaultQueueID = types.StringValue(pool.DefaultQueueID.String())

	if pool.BaseJobTemplate != nil {
		var builder strings.Builder
		encoder := json.NewEncoder(&builder)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(pool.BaseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to serialize Base Job Template",
				fmt.Sprintf("Failed to serialize Base Job Template as JSON string: %s", err),
			)

			return
		}

		model.BaseJobTemplate = types.StringValue(strings.TrimSuffix(builder.String(), "\n"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model WorkPoolResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.Get(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing work pool state",
			fmt.Sprintf("Could not get work pool, unexpected error: %s", err),
		)

		return
	}

	model.ID = types.StringValue(pool.ID.String())
	model.Name = types.StringValue(pool.Name)

	if pool.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(pool.Created.Format(time.RFC3339))
	}

	if pool.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(pool.Updated.Format(time.RFC3339))
	}

	model.Description = types.StringPointerValue(pool.Description)
	model.Type = types.StringValue(pool.Type)
	model.Paused = types.BoolValue(pool.IsPaused)
	model.ConcurrencyLimit = types.Int64PointerValue(pool.ConcurrencyLimit)
	model.DefaultQueueID = types.StringValue(pool.DefaultQueueID.String())

	if pool.BaseJobTemplate != nil {
		var builder strings.Builder
		encoder := json.NewEncoder(&builder)
		err := encoder.Encode(pool.BaseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to serialize Base Job Template",
				fmt.Sprintf("Failed to serialize Base Job Template as JSON string: %s", err),
			)

			return
		}

		model.BaseJobTemplate = types.StringValue(builder.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model WorkPoolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseJobTemplate := map[string]interface{}{}
	if !model.BaseJobTemplate.IsNull() {
		reader := strings.NewReader(model.BaseJobTemplate.ValueString())
		decoder := json.NewDecoder(reader)
		err := decoder.Decode(&baseJobTemplate)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("base_job_template"),
				"Failed to deserialize Base Job Template",
				fmt.Sprintf("Failed to deserialize Base Job Template as JSON object: %s", err),
			)

			return
		}
	}

	err := r.client.Update(ctx, model.Name.ValueString(), api.WorkPoolUpdate{
		Description:      model.Description.ValueStringPointer(),
		IsPaused:         model.Paused.ValueBoolPointer(),
		BaseJobTemplate:  baseJobTemplate,
		ConcurrencyLimit: model.ConcurrencyLimit.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating work pool",
			fmt.Sprintf("Could not update work pool, unexpected error: %s", err),
		)

		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model WorkPoolResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting work pool",
			fmt.Sprintf("Could not delete work pool, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
