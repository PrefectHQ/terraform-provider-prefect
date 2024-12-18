package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/go-test/deep"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

// ObjectsEqual checks to see if two objects are equivalent, accounting for
// differences in the order of the contents.
func ObjectsEqual(obj1, obj2 interface{}) (bool, []string) {
	differences := deep.Equal(obj1, obj2)
	if len(differences) != 0 {
		return false, differences
	}

	return true, nil
}

// CompareNormalizedJSON compares two JSON objects, normalizing them before
// comparison.
//
// This is useful in tests, where we need to mimic the normalization
// that Terraform does to attributes of type jsontypes.Normalized.
func CompareNormalizedJSON(expected, actual map[string]interface{}) error {
	normalizedExpected, err := mapToNormalizedJSON(expected)
	if err != nil {
		return fmt.Errorf("error marshalling expected: %w", err)
	}

	normalizedActual, err := mapToNormalizedJSON(actual)
	if err != nil {
		return fmt.Errorf("error marshalling actual: %w", err)
	}

	if !normalizedActual.Equal(normalizedExpected) {
		return fmt.Errorf("mismatch of JSON objects: expected %s, got %s", normalizedExpected, normalizedActual)
	}

	return nil
}

func mapToNormalizedJSON(m map[string]interface{}) (jsontypes.Normalized, error) {
	jsonByteSlice, err := json.Marshal(m)
	if err != nil {
		return jsontypes.Normalized{}, fmt.Errorf("error marshalling map to JSON: %w", err)
	}

	normalizedJSON := jsontypes.NewNormalizedValue(string(jsonByteSlice))

	return normalizedJSON, nil
}
