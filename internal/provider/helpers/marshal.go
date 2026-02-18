package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// UnmarshalOptional attempts to unmarshal an optional attribute. If it is null or unknown, it returns
// nil and no diagnostics. If it is set, it attempts to unmarshal it and returns any diagnostics.
// Returning nil (rather than an empty map) ensures that struct fields tagged with `omitempty` are
// omitted from JSON serialization, which avoids sending empty `{}` values in PATCH payloads.
func UnmarshalOptional(attribute jsontypes.Normalized) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if attribute.IsNull() || attribute.IsUnknown() {
		return nil, diags
	}
	var result map[string]any
	diags = attribute.Unmarshal(&result)

	return result, diags
}
