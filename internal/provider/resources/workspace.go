package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var (
	_ = resource.ResourceWithConfigure(&WorkspaceResource{})
	_ = resource.ResourceWithImportState(&WorkspaceResource{})
)

// WorkspaceResource contains state for the resource.
type WorkspaceResource struct {
	client api.PrefectClient
}

// WorkspaceResourceModel defines the Terraform resource model.
type WorkspaceResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Created   types.String `tfsdk:"created"`
	Updated   types.String `tfsdk:"updated"`
	AccountID types.String `tfsdk:"account_id"`

	Name        types.String `tfsdk:"name"`
	Handle      types.String `tfsdk:"handle"`
	Description types.String `tfsdk:"description"`
}

// NewWorkspaceResource returns a new WorkspaceResource.
//
//nolint:ireturn // required by Terraform API
func NewWorkspaceResource() resource.Resource {
	return &WorkspaceResource{}
}

// Metadata returns the resource type name.
func (r *WorkspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

// Configure initializes runtime state for the resource.
func (r *WorkspaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Schema defines the schema for the resource.
func (r *WorkspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource representing a Prefect Workspace",
		Version:     0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Workspace UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time of the workspace creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time that the workspace was last updated in RFC 3339 format",
			},
			"account_id": schema.StringAttribute{
				Description: "Account UUID, defaults to the account set in the provider",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the workspace",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "Unique handle for the workspace",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description for the workspace",
				Required:    true,
			},
		},
	}
}

// copyWorkspaceToModel copies an api.Workspace to a WorkspaceResourceModel.
func copyWorkspaceToModel(ctx context.Context, workspace *api.Workspace, model *WorkspaceResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(workspace.ID.String())

	if workspace.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(workspace.Created.Format(time.RFC3339))
	}

	if workspace.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(workspace.Updated.Format(time.RFC3339))
	}

	model.Name = types.StringValue(workspace.Name)
	model.Handle = types.StringValue(workspace.Handle)
	model.Description = types.StringValue(*workspace.Description)

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *WorkspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model WorkspaceResourceModel

	// Populate the model from resource configuration and emit diagnostics on error
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := uuid.Nil
	if !model.AccountID.IsNull() && model.AccountID.ValueString() != "" {
		var err error
		accountID, err = uuid.Parse(model.AccountID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("account_id"),
				"Error parsing Account ID",
				fmt.Sprintf("Could not parse account ID to UUID, unexpected error: %s", err.Error()),
			)

			return
		}
	}

	client, err := r.client.Workspaces(accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	workspace, err := client.Create(ctx, api.WorkspaceCreate{
		Name:        model.Name.ValueString(),
		Handle:      model.Handle.ValueString(),
		Description: model.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace",
			fmt.Sprintf("Could not create workspace, unexpected error: %s", err),
		)

		return
	}

	model.ID = types.StringValue(workspace.ID.String())

	if workspace.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(workspace.Created.Format(time.RFC3339))
	}

	if workspace.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(workspace.Updated.Format(time.RFC3339))
	}

	model.Name = types.StringValue(workspace.Name)
	model.Handle = types.StringValue(workspace.Handle)
	model.Description = types.StringPointerValue(workspace.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *WorkspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model WorkspaceResourceModel

	// Populate the model from state and emit diagnostics on error
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace ID",
			fmt.Sprintf("Could not parse Workspace ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	client, err := r.client.Workspaces(uuid.Nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	workspace, err := client.Get(ctx, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error refreshing Workspace state",
			fmt.Sprintf("Could not read Workspace, unexpected error: %s", err.Error()),
		)

		return
	}

	model.ID = types.StringValue(workspace.ID.String())

	if workspace.Created == nil {
		model.Created = types.StringNull()
	} else {
		model.Created = types.StringValue(workspace.Created.Format(time.RFC3339))
	}

	if workspace.Updated == nil {
		model.Updated = types.StringNull()
	} else {
		model.Updated = types.StringValue(workspace.Updated.Format(time.RFC3339))
	}

	model.Name = types.StringValue(workspace.Name)
	model.Handle = types.StringValue(workspace.Handle)
	model.Description = types.StringPointerValue(workspace.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *WorkspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model WorkspaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace ID",
			fmt.Sprintf("Could not parse Workspace ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	client, err := r.client.Workspaces(uuid.Nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	err = client.Update(ctx, workspaceID, api.WorkspaceUpdate{
		Name:        model.Name.ValueStringPointer(),
		Handle:      model.Handle.ValueStringPointer(),
		Description: model.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating workspace",
			fmt.Sprintf("Could not update workspace, unexpected error: %s", err),
		)

		return
	}

	workspace, _ := client.Get(ctx, workspaceID)
	model.Created = types.StringValue(workspace.Created.Format(time.RFC3339))
	model.Updated = types.StringValue(workspace.Created.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *WorkspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model WorkspaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Error parsing Workspace ID",
			fmt.Sprintf("Could not parse Workspace ID to UUID, unexpected error: %s", err.Error()),
		)

		return
	}

	client, err := r.client.Workspaces(uuid.Nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace client",
			fmt.Sprintf("Could not create workspace client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", err.Error()),
		)
	}

	err = client.Delete(ctx, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Workspace",
			fmt.Sprintf("Could not delete Workspace, unexpected error: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *WorkspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if strings.HasPrefix(req.ID, "name/") {
		name := strings.TrimPrefix(req.ID, "name/")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}
