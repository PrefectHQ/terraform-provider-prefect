package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stringNotNullValidator struct {
	tfsdk.AttributeValidator
}

func (v stringNotNullValidator) Description(ctx context.Context) string {
	return "string must not be null"
}

func (v stringNotNullValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringNotNullValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Null {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			req.AttributePath.String()+" is null",
			"must not be null",
		)
	}
}

func StringNotNull() tfsdk.AttributeValidator {
	return &stringNotNullValidator{}
}
