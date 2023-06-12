package customtypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ = basetypes.StringValuable(&TimestampValue{})
	_ = basetypes.StringValuableWithSemanticEquals(&TimestampValue{})
	_ = fmt.Stringer(&TimestampValue{})
)

type TimestampValue struct {
	basetypes.StringValue
}

// NewTimestampNull creates a Timestamp with a null value. Determine
// whether the value is null via the Timestamp type IsNull method.
func NewTimestampNull() TimestampValue {
	return TimestampValue{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewTimestampUnknown creates a Timestamp with an unknown value.
// Determine whether the value is null via the Timestamp type IsNull
// method.
func NewTimestampUnknown() TimestampValue {
	return TimestampValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewTimestampValue creates a Timestamp with a known value. Access
// the value via the TimestampValue type ValueTime method.
func NewTimestampValue(value time.Time) TimestampValue {
	return TimestampValue{
		StringValue: basetypes.NewStringValue(value.Format(time.RFC3339)),
	}
}

// NewTimestampPointerValue creates a Timestamp with a null value if
// nil or a known value. Access the value via the TimestampValue type
// ValueTimePointer method.
func NewTimestampPointerValue(value *time.Time) TimestampValue {
	if value == nil {
		return NewTimestampNull()
	}

	return NewTimestampValue(*value)
}

// Equal returns true if this timestamp is equal to o.
func (v TimestampValue) Equal(o attr.Value) bool {
	other, ok := o.(TimestampValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// Type returns an instance of the type.
//
//nolint:ireturn // required to implement StringValuable
func (v TimestampValue) Type(_ context.Context) attr.Type {
	return TimestampType{}
}

func (v TimestampValue) String() string {
	return "TimestampValue"
}

// StringSemanticEquals checks if two TimestampValue objects have
// equivalent values, even if they are not equal.
func (v TimestampValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TimestampValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected value type %T but got value type %T. Please report this to the provider developers.", v, newValuable),
		)

		return false, diags
	}

	priorTime := v.ValueTime()
	newTime := newValue.ValueTime()

	return priorTime.Equal(newTime), nil
}

// ValueTime returns the timestamp as a time.Time.
func (v TimestampValue) ValueTime() time.Time {
	value, _ := time.Parse(time.RFC3339, v.StringValue.ValueString())

	return value
}
