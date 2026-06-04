package helpers_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

func TestIs404Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    error
		expected bool
	}{
		{
			name:     "http error encoding status_code=404",
			input:    fmt.Errorf("http error: status_code=404, error=not found, body="),
			expected: true,
		},
		{
			name: "synthetic not-found wrapping ErrNotFound (workspace_access shape)",
			input: fmt.Errorf("workspace access not found for accessID: %s: %w",
				"be2d9981-300a-4c66-a4f7-7ed298a8830c", helpers.ErrNotFound),
			expected: true,
		},
		{
			name: "synthetic not-found wrapping ErrNotFound (team_access shape)",
			input: fmt.Errorf("client.Read: team access not found for member ID: %s: %w",
				"be2d9981-300a-4c66-a4f7-7ed298a8830c", helpers.ErrNotFound),
			expected: true,
		},
		{
			name:     "bare ErrNotFound sentinel",
			input:    helpers.ErrNotFound,
			expected: true,
		},
		{
			name:     "deeply wrapped ErrNotFound",
			input:    fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", helpers.ErrNotFound)),
			expected: true,
		},
		{
			name:     "unrelated server error",
			input:    fmt.Errorf("http error: status_code=500, error=boom, body="),
			expected: false,
		},
		{
			name:     "generic error without status code or sentinel",
			input:    errors.New("something else went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := helpers.Is404Error(tt.input); got != tt.expected {
				t.Errorf("Is404Error(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
