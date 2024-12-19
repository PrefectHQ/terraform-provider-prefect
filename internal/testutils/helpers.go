package testutils

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// MapToNormalizedJSON converts a map[string]interface{} to a jsontypes.Normalized.
// This is useful in tests, where we need to mimic the normalization
// that Terraform does to attributes of type jsontypes.Normalized.
func MapToNormalizedJSON(m map[string]interface{}) (jsontypes.Normalized, error) {
	jsonByteSlice, err := json.Marshal(m)
	if err != nil {
		return jsontypes.Normalized{}, fmt.Errorf("error marshalling map to JSON: %w", err)
	}

	normalizedJSON := jsontypes.NewNormalizedValue(string(jsonByteSlice))

	return normalizedJSON, nil
}

// NormalizedValueForJSON generates the normalized JSON value for a map[string]interface{}
// for use as the expected value in the test.
//
// Mimics the normalization that Terraform does to attributes of type jsontypes.Normalized.
//
// Requires unmarshaling to a map[string]interface{} before normalization.
func NormalizedValueForJSON(jsonValue string) (string, error) {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonValue), &jsonMap); err != nil {
		return "", fmt.Errorf("error unmarshalling template: %w", err)
	}

	normalizedJSON, err := MapToNormalizedJSON(jsonMap)
	if err != nil {
		return "", fmt.Errorf("error marshalling jsonMap to normalized JSON: %w", err)
	}

	return normalizedJSON.ValueString(), nil
}
