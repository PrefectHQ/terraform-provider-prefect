package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// UnmarshalOptional attempts to unmarshal an optional attribute. If it is null, it returns an empty
// map[string]interface{} and no diagnostics. If it is not null, it attempts to unmarshal it and returns
// any diagnostics.
func UnmarshalOptional(attribute jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := map[string]interface{}{}
	if !attribute.IsNull() {
		diags = attribute.Unmarshal(&result)
	}

	return result, diags
}
