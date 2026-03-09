package resources_test

import (
	"testing"

	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

func TestIsEmptyJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "empty object", input: "{}", expected: true},
		{name: "empty object with whitespace", input: "{ }", expected: true},
		{name: "non-empty object", input: `{"type":"object"}`, expected: false},
		{name: "nested object", input: `{"properties":{"name":{"type":"string"}}}`, expected: false},
		{name: "invalid JSON", input: "not-json", expected: false},
		{name: "empty string", input: "", expected: false},
		{name: "JSON array", input: "[]", expected: false},
		{name: "JSON null", input: "null", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resources.IsEmptyJSONObject(tt.input)
			if got != tt.expected {
				t.Errorf("IsEmptyJSONObject(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
