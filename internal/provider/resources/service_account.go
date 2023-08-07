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

var (
	_ = resource.ResourceWithConfigure(&ServiceAccountResource{})
	_ = resource.ResourceWithImportState(&ServiceAccountResource{})
)

type ServiceAccountResource struct {
	client api.PrefectClient
}

type ServiceAccountResourceModel struct {
	ID             types.String               `tfsdk:"id"`
	Name           types.String               `tfsdk:"name"`
	Created        customtypes.TimestampValue `tfsdk:"created"`
	Updated        customtypes.TimestampValue `tfsdk:"updated"`
	AccountID      customtypes.UUIDValue      `tfsdk:"account_id"`

	RoleName       types.String               `tfsdk:"account_role_name"`
	APIKeyID       types.String               `tfsdk:"api_key_id"`
	APIKeyName     types.String               `tfsdk:"api_key_name"`
	APIKeyCreated  customtypes.TimestampValue `tfsdk:"api_key_created"`
	APIKeyExpires  customtypes.TimestampValue `tfsdk:"api_key_expiration"`
	APIKey         types.String               `tfsdk:"api_key"`
}

func NewServiceAccountResource() resource.Resource {
	return &ServiceAccountResource{}
}

func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Resource representing a Prefect service account",
		Version:     1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Service account UUID",
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Name of the service account",
			},
			"created": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time of the service account creation in RFC 3339 format",
			},
			"updated": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time that the service account was last updated in RFC 3339 format",
			},
			"account_id": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.UUIDType{},
				Description: "Account UUID, defaults to the account set in the provider",
			},
			"account_role_name": schema.StringAttribute{
				Computed: true,
				Description: "Role name of the service account",
			},
			"api_key_id": schema.StringAttribute{
				Computed: true,
				Description: "API Key ID associated with the service account",
			},
			"api_key_name": schema.StringAttribute{
				Computed: true,
				Description: "API Key Name associated with the service account",
			},
			"api_key_created": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time that the API Key was created in RFC 3339 format",
			},
			"api_key_expiration": schema.StringAttribute{
				Computed: true,
				CustomType: customtypes.TimestampType{},
				Description: "Date and time that the API Key expires in RFC 3339 format",
			},
			"api_key": schema.StringAttribute{
				Computed: true,
				Description: "API Key associated with the service account",
			},
		},
	}
}

// @TODO Implement CRUD operations for tfsdk