package customtypes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ = basetypes.StringValuable(&UUIDValue{})
	_ = basetypes.StringValuableWithSemanticEquals(&UUIDValue{})
	_ = fmt.Stringer(&UUIDValue{})
)

// UUIDValue implements a custom Terraform value that represents
// a valid UUID.
type UUIDValue struct {
	basetypes.StringValue
}

// NewUUIDNull creates a UUID with a null value. Determine
// whether the value is null via the UUID type IsNull method.
func NewUUIDNull() UUIDValue {
	return UUIDValue{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewUUIDUnknown creates a UUID with an unknown value.
// Determine whether the value is null via the UUID type IsNull
// method.
func NewUUIDUnknown() UUIDValue {
	return UUIDValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewUUIDValue creates a UUID with a known value. Access
// the value via the UUIDValue type ValueUUID method.
func NewUUIDValue(value uuid.UUID) UUIDValue {
	return UUIDValue{
		StringValue: basetypes.NewStringValue(value.String()),
	}
}

// NewUUIDPointerValue creates a UUID with a null value if
// nil or a known value. Access the value via the UUIDValue type
// ValueUUIDPointer method.
func NewUUIDPointerValue(value *uuid.UUID) UUIDValue {
	if value == nil {
		return NewUUIDNull()
	}

	return NewUUIDValue(*value)
}

// Equal returns true if this timestamp is equal to o.
func (v UUIDValue) Equal(o attr.Value) bool {
	other, ok := o.(UUIDValue)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// Type returns an instance of the type.
//
//nolint:ireturn // required to implement StringValuable
func (v UUIDValue) Type(_ context.Context) attr.Type {
	return UUIDType{}
}

func (v UUIDValue) String() string {
	return "UUIDValue"
}

// StringSemanticEquals checks if two UUIDValue objects have
// equivalent values, even if they are not equal.
func (v UUIDValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(UUIDValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected value type %T but got value type %T. Please report this to the provider developers.", v, newValuable),
		)

		return false, diags
	}

	priorUUID := v.ValueUUID()
	newUUID := newValue.ValueUUID()

	return priorUUID == newUUID, nil
}

// ValueUUID returns the UUID as a uuid.UUID. If the value
// is unknown or null, this will return uuid.Nil.
func (v UUIDValue) ValueUUID() uuid.UUID {
	if v.IsNull() || v.IsUnknown() {
		return uuid.Nil
	}

	value, _ := uuid.Parse(v.StringValue.ValueString())

	return value
}

// ValueUUIDPointer returns the UUID as a *uuid.UUID. If the value
// is unknown or null, this will return nil.
func (v UUIDValue) ValueUUIDPointer() *uuid.UUID {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	value := v.ValueUUID()

	return &value
}
