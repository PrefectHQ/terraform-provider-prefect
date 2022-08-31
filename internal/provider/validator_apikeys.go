package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type apiKeyNamesAreUniqueValidator struct {
	tfsdk.AttributeValidator
}

func (v apiKeyNamesAreUniqueValidator) Description(ctx context.Context) string {
	return "each api key must have a unique name"
}

func (v apiKeyNamesAreUniqueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v apiKeyNamesAreUniqueValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var apiKeys []apiKey
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &apiKeys)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	seen := map[string]bool{}

	for _, apiKey := range apiKeys {
		name := apiKey.Name.Value

		if _, ok := seen[name]; ok {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"api key name is not unique",
				fmt.Sprintf("More than one api key with name %s", name),
			)
		}
		seen[name] = true
	}
}

func APIKeyNamesAreUnique() tfsdk.AttributeValidator {
	return &apiKeyNamesAreUniqueValidator{}
}
