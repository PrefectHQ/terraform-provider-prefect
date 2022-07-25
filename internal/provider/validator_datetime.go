package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type isRFC3339UTCValidator struct {
	tfsdk.AttributeValidator
}

func (v isRFC3339UTCValidator) Description(ctx context.Context) string {
	return "must be a UTC date time in RFC3339 format, eg: 2015-10-21T13:00:00+00:00 to match the Prefect API and avoid changes on refresh."
}

func (v isRFC3339UTCValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v isRFC3339UTCValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if !str.Null {
		if _, err := time.Parse(time.RFC3339, str.Value); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				req.AttributePath.String()+" is not a valid timestamp",
				fmt.Sprintf("%v. Expected a valid RFC3339 timestamp.", err),
			)
		}
		if !strings.HasSuffix(str.Value, "+00:00") {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				req.AttributePath.String()+" has unsupported timezone",
				fmt.Sprintf("%s must be in the +00:00 timezone.", str.Value),
			)
		}
	}
}

func IsRFC3339Time() tfsdk.AttributeValidator {
	return &isRFC3339UTCValidator{}
}
