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

// NormalizeParameterOpenAPISchema normalizes a parameter OpenAPI schema to match
// the format expected by Prefect OSS (and Cloud). When an empty object {} is provided,
// it converts it to a valid OpenAPI schema: {"type": "object", "properties": {}}.
// This ensures consistency between what the provider sends and what the API returns,
// preventing "inconsistent result after apply" errors.
func NormalizeParameterOpenAPISchema(schema map[string]interface{}) map[string]interface{} {
	// If the schema is empty, return the normalized empty schema
	if len(schema) == 0 {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	// If the schema is missing "type" or "properties", add them
	if _, hasType := schema["type"]; !hasType {
		schema["type"] = "object"
	}
	if _, hasProperties := schema["properties"]; !hasProperties {
		schema["properties"] = map[string]interface{}{}
	}

	return schema
}
