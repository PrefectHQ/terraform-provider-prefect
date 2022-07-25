package util

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToString(s types.String) *string {
	if s.Null {
		return nil
	}
	if s.Unknown {
		return nil
	}
	return &s.Value
}

func FromString(s *string) types.String {
	if s == nil {
		return types.String{Null: true}
	}
	return types.String{Value: *s}
}
