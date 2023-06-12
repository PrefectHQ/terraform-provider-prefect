package customtypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ = basetypes.StringTypable(&TimestampType{})
	_ = xattr.TypeWithValidate(&TimestampType{})
	_ = fmt.Stringer(&TimestampType{})
)

type TimestampType struct {
	basetypes.StringType
}

func (t TimestampType) Equal(o attr.Type) bool {
	other, ok := o.(TimestampType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TimestampType) String() string {
	return "TimestampType"
}

// ValueFromString converts a string value to a TimestampValue.
//
//nolint:ireturn // required to implement StringTypable
func (t TimestampType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := TimestampValue{
		StringValue: in,
	}

	return value, nil
}

// ValueFromTerraform converts a Terraform value to a TimestampValue.
//
//nolint:ireturn // required to implement StringTypable
func (t TimestampType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

//nolint:ireturn // required to implement StringTypable
func (t TimestampType) ValueType(_ context.Context) attr.Value {
	return TimestampValue{}
}

func (t TimestampType) Validate(_ context.Context, value tftypes.Value, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	var diags diag.Diagnostics
	var timestampStr string
	if err := value.As(&timestampStr); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid Terraform Value",
			fmt.Sprintf("Failed to convert %T to string: %s. Please report this issue to the provider developers.", value, err.Error()),
		)

		return diags
	}

	if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
		diags.AddAttributeError(
			valuePath,
			"Invalid RFC 3339 String Value",
			fmt.Sprintf("Failed to parse string %q as RFC 3339 timestamp (YYYY-MM-DDTHH:MM:SSZ): %s", timestampStr, err.Error()),
		)

		return diags
	}

	return diags
}
