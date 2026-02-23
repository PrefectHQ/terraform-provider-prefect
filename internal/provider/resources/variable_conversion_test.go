package resources // nolint:testpackage // need access to private convertAPIValueToDynamic function

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertAPIValueToDynamic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		input         any
		expectedType  string
		expectedValue any
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
			input:        []any{"foo", "bar"},
			expectedType: "types.Tuple",
			expectError:  false,
		},
		{
			name:         "object value",
			input:        map[string]any{"key": "value", "number": float64(42)},
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

func TestConvertAttrValueToNative(t *testing.T) {
	t.Parallel()

	t.Run("string value", func(t *testing.T) {
		t.Parallel()
		result, diags := convertAttrValueToNative(types.StringValue("hello"))
		require.False(t, diags.HasError())
		assert.Equal(t, "hello", result)
	})

	t.Run("JSON string value is preserved as-is", func(t *testing.T) {
		t.Parallel()
		result, diags := convertAttrValueToNative(types.StringValue(`{"name":"dev"}`))
		require.False(t, diags.HasError())
		assert.Equal(t, `{"name":"dev"}`, result)
	})

	t.Run("number value", func(t *testing.T) {
		t.Parallel()
		result, diags := convertAttrValueToNative(types.NumberValue(big.NewFloat(42)))
		require.False(t, diags.HasError())
		assert.Equal(t, float64(42), result)
	})

	t.Run("boolean value", func(t *testing.T) {
		t.Parallel()
		result, diags := convertAttrValueToNative(types.BoolValue(true))
		require.False(t, diags.HasError())
		assert.Equal(t, true, result)
	})

	t.Run("tuple of strings", func(t *testing.T) {
		t.Parallel()
		tupleVal, tupleDiags := types.TupleValue(
			[]attr.Type{types.StringType, types.StringType},
			[]attr.Value{types.StringValue("foo"), types.StringValue("bar")},
		)
		require.False(t, tupleDiags.HasError())

		result, diags := convertAttrValueToNative(tupleVal)
		require.False(t, diags.HasError())
		assert.Equal(t, []any{"foo", "bar"}, result)
	})

	t.Run("tuple of mixed types", func(t *testing.T) {
		t.Parallel()
		tupleVal, tupleDiags := types.TupleValue(
			[]attr.Type{types.StringType, types.NumberType, types.BoolType},
			[]attr.Value{
				types.StringValue("foo"),
				types.NumberValue(big.NewFloat(123)),
				types.BoolValue(true),
			},
		)
		require.False(t, tupleDiags.HasError())

		result, diags := convertAttrValueToNative(tupleVal)
		require.False(t, diags.HasError())
		assert.Equal(t, []any{"foo", float64(123), true}, result)
	})

	t.Run("object value", func(t *testing.T) {
		t.Parallel()
		objVal, objDiags := types.ObjectValue(
			map[string]attr.Type{"key": types.StringType, "number": types.NumberType},
			map[string]attr.Value{
				"key":    types.StringValue("value"),
				"number": types.NumberValue(big.NewFloat(42)),
			},
		)
		require.False(t, objDiags.HasError())

		result, diags := convertAttrValueToNative(objVal)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]any{"key": "value", "number": float64(42)}, result)
	})

	t.Run("nested object containing tuple", func(t *testing.T) {
		t.Parallel()
		tupleVal, tupleDiags := types.TupleValue(
			[]attr.Type{types.StringType, types.StringType},
			[]attr.Value{types.StringValue("a"), types.StringValue("b")},
		)
		require.False(t, tupleDiags.HasError())

		objVal, objDiags := types.ObjectValue(
			map[string]attr.Type{
				"name": types.StringType,
				"list": tupleVal.Type(context.Background()),
			},
			map[string]attr.Value{
				"name": types.StringValue("test"),
				"list": tupleVal,
			},
		)
		require.False(t, objDiags.HasError())

		result, diags := convertAttrValueToNative(objVal)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]any{
			"name": "test",
			"list": []any{"a", "b"},
		}, result)
	})
}
