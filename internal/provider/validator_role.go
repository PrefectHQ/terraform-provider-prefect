package provider

import (
	"context"
	"fmt"
	"terraform-provider-prefect/api"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleValidator struct {
	tfsdk.AttributeValidator
}

func (v roleValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("role must be one of [%s, %s, %s]", api.Membership_roleReadOnlyUser, api.Membership_roleTenantAdmin, api.Membership_roleUser)
}

func (v roleValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v roleValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var role types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &role)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	switch role.Value {
	case
		api.Membership_roleReadOnlyUser,
		api.Membership_roleTenantAdmin,
		api.Membership_roleUser:
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.AttributePath,
		req.AttributePath.String()+" is an invalid role",
		v.Description(ctx),
	)
}

func RoleIsValid() tfsdk.AttributeValidator {
	return &roleValidator{}
}
