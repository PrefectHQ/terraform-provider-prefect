package provider

import (
	"context"
	"fmt"

	"terraform-provider-prefect/api"
	"terraform-provider-prefect/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectResourceType struct{}

// project resource schema
func (r projectResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Project resource.
`,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Server generated UUID",
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					StringNotNull(),
				},
				Description: "Name",
			},
			"description": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Description",
			},
		},
	}, nil
}

// New resource instance
func (r projectResourceType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return projectResource{
		provider: provider,
	}, diags
}

type projectResource struct {
	provider provider
}

type projectResourceData struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// Create a new resource
func (r projectResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource.",
		)
		return
	}

	// Retrieve values from config (ie: .tf file)
	var config projectResourceData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new project
	project, err := api.CreateProject(r.provider.client.GQLClient, ctx, config.Name.Value, util.ToString(config.Description))
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create project",
			fmt.Sprintf("%v", err),
		)
		return
	}

	// Set values that were unknown
	config.ID = types.String{Value: string(*project.Id)}

	// Set state
	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// Read resource information
func (r projectResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state projectResourceData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get project from API
	project, err := api.GetProject(r.provider.client.GQLClient, ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read project",
			fmt.Sprintf("%v (project ID = %s)", err, state.ID.Value),
		)
		return
	}

	if project == nil {
		// project doesn't exist ie: has been deleted outside of terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with values from API response
	state.Name = types.String{Value: project.Name}
	state.Description = util.FromString(project.Description)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource
func (r projectResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan projectResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state projectResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update project name by calling api
	if !state.Name.Equal(plan.Name) {
		_, err := api.SetProjectName(r.provider.client.GQLClient, ctx, (api.UUID)(state.ID.Value), plan.Name.Value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not update project name",
				fmt.Sprintf("%v (project ID = %s)", err, state.ID.Value),
			)
			return
		}
		state.Name = plan.Name
	}

	// update project description by calling api
	if !state.Description.Equal(plan.Description) {
		_, err := api.SetProjectDescription(r.provider.client.GQLClient, ctx, (api.UUID)(state.ID.Value), util.ToString(plan.Description))
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not update project description",
				fmt.Sprintf("%v (project ID = %s)", err, state.ID.Value),
			)
			return
		}
		state.Description = plan.Description
	}

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Delete resource
func (r projectResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state projectResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete project by calling API
	successPayload, err := api.DeleteProject(r.provider.client.GQLClient, ctx, (api.UUID)(state.ID.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete project",
			fmt.Sprintf("%v (project ID = %s)", err, state.ID.Value),
		)
		return
	}
	if !*successPayload.Success {
		resp.Diagnostics.AddError(
			"Could not delete project",
			fmt.Sprintf("%v (project ID = %s)", *successPayload.Error, state.ID.Value),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r projectResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
