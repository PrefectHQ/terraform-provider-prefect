package testutils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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

// expectKnownValue is a helper function that creates a statecheck.StateCheck
// that can be used to check the known value of a resource attribute.
//
//nolint:ireturn // required for testing
func expectKnownValue(resourceName, path string, check knownvalue.Check) statecheck.StateCheck {
	pathValue := tfjsonpath.New(path)

	if strings.Contains(path, ".") {
		keys := strings.Split(path, ".")

		pathValue = tfjsonpath.New(keys[0])
		for _, key := range keys[1:] {
			if keyInt, err := strconv.Atoi(key); err == nil {
				pathValue = pathValue.AtSliceIndex(keyInt)
			} else {
				pathValue = pathValue.AtMapKey(key)
			}
		}
	}

	return statecheck.ExpectKnownValue(
		resourceName,
		pathValue,
		check,
	)
}

// ExpectKnownValue returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute.
//
//nolint:ireturn // required for testing
func ExpectKnownValue(resourceName, path, value string) statecheck.StateCheck {
	return expectKnownValue(resourceName, path, knownvalue.StringExact(value))
}

// ExpectKnownValueList returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is a list of strings.
//
//nolint:ireturn // required for testing
func ExpectKnownValueList(resourceName, path string, values []string) statecheck.StateCheck {
	knownValueChecks := []knownvalue.Check{}
	for _, value := range values {
		knownValueChecks = append(knownValueChecks, knownvalue.StringExact(value))
	}

	return expectKnownValue(resourceName, path, knownvalue.ListExact(knownValueChecks))
}

// ExpectKnownValueBool returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is a boolean.
//
//nolint:ireturn // required for testing
func ExpectKnownValueBool(resourceName, path string, value bool) statecheck.StateCheck {
	return expectKnownValue(resourceName, path, knownvalue.Bool(value))
}

// ExpectKnownValueNotNull returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is not null.
//
//nolint:ireturn // required for testing
func ExpectKnownValueNotNull(resourceName, path string) statecheck.StateCheck {
	return expectKnownValue(resourceName, path, knownvalue.NotNull())
}

// ExpectKnownValueNumber returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is a number.
//
//nolint:ireturn // required for testing
func ExpectKnownValueNumber(resourceName, path string, value int64) statecheck.StateCheck {
	return expectKnownValue(resourceName, path, knownvalue.Int64Exact(value))
}

// ExpectKnownValueFloat returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is a float.
//
//nolint:ireturn // required for testing
func ExpectKnownValueFloat(resourceName, path string, value float64) statecheck.StateCheck {
	return expectKnownValue(resourceName, path, knownvalue.Float64Exact(value))
}

// ExpectKnownValueMap returns a statecheck.StateCheck that can be used to
// check the known value of a resource attribute that is a map.
//
//nolint:ireturn // required for testing
func ExpectKnownValueMap(resourceName, path string, value map[string]string) statecheck.StateCheck {
	knownValueChecks := map[string]knownvalue.Check{}
	for k, v := range value {
		knownValueChecks[k] = knownvalue.StringExact(v)
	}

	return expectKnownValue(resourceName, path, knownvalue.MapExact(knownValueChecks))
}

// CompareValuePairs is a helper function that creates a statecheck.StateCheck
// that can be used to check the known value of a resource attribute.
//
//nolint:ireturn // required for testing
func CompareValuePairs(resourceName1, path1, resourceName2, path2 string) statecheck.StateCheck {
	return statecheck.CompareValuePairs(
		resourceName1,
		tfjsonpath.New(path1),
		resourceName2,
		tfjsonpath.New(path2),
		compare.ValuesSame(),
	)
}
