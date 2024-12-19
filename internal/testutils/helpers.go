package testutils

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// convertToNormalizedJSON converts an interface to a jsontypes.Normalized.
// This is useful in tests, where we need to mimic the normalization
// that Terraform does to attributes of type jsontypes.Normalized.
func convertToNormalizedJSON(i interface{}) (jsontypes.Normalized, error) {
	jsonByteSlice, err := json.Marshal(i)
	if err != nil {
		return jsontypes.Normalized{}, fmt.Errorf("error marshalling interface to JSON: %w", err)
	}

	normalizedJSON := jsontypes.NewNormalizedValue(string(jsonByteSlice))

	return normalizedJSON, nil
}

// NormalizedValueForJSON generates the normalized JSON string value for an
// interface for use as the expected value in the test.
//
// Mimics the normalization that Terraform does to attributes of type
// jsontypes.Normalized.
//
// Requires unmarshaling to a interface before normalization.
func NormalizedValueForJSON(t *testing.T, jsonValue string) string {
	t.Helper()

	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonValue), &jsonObj); err != nil {
		t.Fatalf("error unmarshalling JSON value %s: %s", jsonValue, err.Error())
	}

	normalizedJSON, err := convertToNormalizedJSON(jsonObj)
	if err != nil {
		t.Fatalf("error converting JSON value %s to normalized JSON: %s", jsonValue, err.Error())
	}

	return normalizedJSON.ValueString()
}
