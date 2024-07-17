package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ = basetypes.StringTypable(&UUIDType{})
	_ = fmt.Stringer(&UUIDType{})
)

// UUIDType implements a custom Terraform type that represents
// a valid UUID.
type UUIDType struct {
	basetypes.StringType
}

// Equal returns true of this UUID and o are equal.
func (t UUIDType) Equal(o attr.Type) bool {
	other, ok := o.(UUIDType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// String represents a string representation of UUIDType.
func (t UUIDType) String() string {
	return "UUIDType"
}

// ValueFromString converts a string value to a UUIDValue.
//
//nolint:ireturn // required to implement StringTypable
func (t UUIDType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := UUIDValue{
		StringValue: in,
	}

	return value, nil
}

// ValueFromTerraform converts a Terraform value to a UUIDValue.
//
//nolint:ireturn // required to implement StringTypable
func (t UUIDType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("unexpected error converting value from Terraform: %w", err)
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

// ValueType returns an instance of the value.
//
//nolint:ireturn // required to implement StringTypable
func (t UUIDType) ValueType(_ context.Context) attr.Value {
	return UUIDValue{}
}
