package customtypes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ = basetypes.StringTypable(&UUIDType{})
	_ = xattr.TypeWithValidate(&UUIDType{})
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

// Validate ensures that the string can be converted to a UUIDValue.
func (t UUIDType) Validate(_ context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var uuidStr string
	if err := value.As(&uuidStr); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid Terraform Value",
			fmt.Sprintf("Failed to convert %T to string: %s. Please report this issue to the provider developers.", value, err.Error()),
		)

		return diags
	}

	if _, err := uuid.Parse(uuidStr); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid UUID String Value",
			fmt.Sprintf("Failed to parse string %q as a UUID: %s", uuidStr, err.Error()),
		)

		return diags
	}

	return diags
}
