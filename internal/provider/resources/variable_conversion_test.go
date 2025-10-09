package resources // nolint:testpackage // need access to private convertAPIValueToDynamic function

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAPIValueToDynamic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		input         interface{}
		expectedType  string
		expectedValue interface{}
		expectError   bool
	}{
		{
			name:          "string value",
			input:         "hello-world",
			expectedType:  "types.String",
			expectedValue: "hello-world",
			expectError:   false,
		},
		{
			name:          "number value",
			input:         float64(123.45),
			expectedType:  "types.Number",
			expectedValue: float64(123.45),
			expectError:   false,
		},
		{
			name:          "boolean value",
			input:         true,
			expectedType:  "types.Bool",
			expectedValue: true,
			expectError:   false,
		},
		{
			name:          "nil value",
			input:         nil,
			expectedType:  "types.Dynamic",
			expectedValue: nil,
			expectError:   false,
		},
		{
			name:         "array value",
			input:        []interface{}{"foo", "bar"},
			expectedType: "types.Tuple",
			expectError:  false,
		},
		{
			name:         "object value",
			input:        map[string]interface{}{"key": "value", "number": float64(42)},
			expectedType: "types.Object",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, diags := convertAPIValueToDynamic(ctx, tt.input)

			if tt.expectError {
				assert.True(t, diags.HasError(), "expected error but got none")

				return
			}

			require.False(t, diags.HasError(), "unexpected error: %v", diags)
			assert.NotNil(t, result, "result should not be nil")

			// Check specific values for simple types
			switch tt.expectedType {
			case "types.String":
				strVal, ok := result.UnderlyingValue().(types.String)
				require.True(t, ok, "expected types.String")
				assert.Equal(t, tt.expectedValue, strVal.ValueString())

			case "types.Number":
				numVal, ok := result.UnderlyingValue().(types.Number)
				require.True(t, ok, "expected types.Number")
				numFloat, _ := numVal.ValueBigFloat().Float64()
				assert.Equal(t, tt.expectedValue, numFloat)

			case "types.Bool":
				boolVal, ok := result.UnderlyingValue().(types.Bool)
				require.True(t, ok, "expected types.Bool")
				assert.Equal(t, tt.expectedValue, boolVal.ValueBool())

			case "types.Dynamic":
				if tt.expectedValue == nil {
					assert.True(t, result.IsNull(), "expected null dynamic value")
				}

			case "types.Tuple":
				tupleVal, ok := result.UnderlyingValue().(types.Tuple)
				require.True(t, ok, "expected types.Tuple")
				assert.NotNil(t, tupleVal)

			case "types.Object":
				objVal, ok := result.UnderlyingValue().(types.Object)
				require.True(t, ok, "expected types.Object")
				assert.NotNil(t, objVal)
			}
		})
	}
}
