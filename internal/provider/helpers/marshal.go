package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SafeUnmarshal is a helper function for safely unmarshalling a JSON object by checking if
// it is null before attempting to unmarshal it. This should always be used for optional attributes.
func SafeUnmarshal(attribute jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := map[string]interface{}{}
	if !attribute.IsNull() {
		diags = attribute.Unmarshal(&result)
	}

	return result, diags
}
