package prefect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

type ServiceAccountResourceModel struct {
	Name            	types.String 			`tfsdk:"name"`
	APIKeyExpiration 	types.String 			`tfsdk:"api_key_expiration"`
	AccountRoleId   	types.String 			`tfsdk:"account_role_id"`
	ID              	types.String 			`tfsdk:"id"`
	AccountID			customtypes.UUIDValue   `tfsdk:"account_id"`
}

type ServiceAccountResourceType struct{}

func (r ServiceAccountResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diagnostics.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"api_key_expiration": {
				Type:     types.StringType,
				Required: true,
			},
			"account_role_id": {
				Type:     types.StringType,
				Required: true,
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (r ServiceAccountResourceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diagnostics.Diagnostics) {
	client := p.(*provider).client

	return &ServiceAccountResource{
		client: client,
	}, nil
}

type ServiceAccountResource struct {
	client api.PrefectClient
}

func (r *ServiceAccountResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan ServiceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := api.ServiceAccountCreate{
		Name: plan.Name.Value,
		APIKeyExpiration: plan.APIKeyExpiration.Value,
		AccountRoleId: plan.AccountRoleId.Value,
	}

	// @TODO: Set values in CreateRequest
	createdAccount, err := r.client.Create(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create service account",
			"An unexpected error was encountered while creating the service account. The service account could not be created.",
		)
		return
	}

	plan.ID = types.String{Value: createdAccount.ID}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var model ServiceAccountResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.client.Get(ctx, model.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read service account",
			"An unexpected error was encountered while reading the service account. The service account could not be read.",
		)
		return
	}

	model.Name = types.String{Value: account.Name}
	model.APIKeyExpiration = types.String{Value: account.APIKey.Expiration}
	model.AccountRoleId = types.String{Value: account.AccountRoleId}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan ServiceAccountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := api.ServiceAccountUpdate{
		Name: plan.Name.Value,
	}

	_, err := r.client.Update(ctx, plan.ID.Value, updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not update service account",
			"An unexpected error was encountered while updating the service account. The service account could not be updated.",
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var model ServiceAccountResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(ctx, model.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete service account",
			"An unexpected error was encountered while deleting the service account. The service account could not be deleted.",
		)
		return
	}
}
