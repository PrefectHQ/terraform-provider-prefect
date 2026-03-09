package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/resources"
)

func TestUseServerValueIfEmpty_UnknownConfig(t *testing.T) {
	t.Parallel()

	req := planmodifier.StringRequest{
		ConfigValue: types.StringUnknown(),
		PlanValue:   types.StringValue("plan"),
		StateValue:  types.StringValue("state"),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != "plan" {
		t.Errorf("expected plan value to remain unchanged, got %q", resp.PlanValue.ValueString())
	}
}

func TestUseServerValueIfEmpty_NullConfig_NullState(t *testing.T) {
	t.Parallel()

	// On create (null state), leave plan as-is so the framework treats it as unknown.
	req := planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to remain unknown on create, got %q", resp.PlanValue)
	}
}

func TestUseServerValueIfEmpty_NullConfig_ExistingState(t *testing.T) {
	t.Parallel()

	// On update with null config, preserve the state value.
	stateVal := types.StringValue(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	req := planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  stateVal,
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != stateVal.ValueString() {
		t.Errorf("expected plan value to match state %q, got %q", stateVal.ValueString(), resp.PlanValue.ValueString())
	}
}

func TestUseServerValueIfEmpty_EmptyObject(t *testing.T) {
	t.Parallel()

	// Config is "{}" — should mark plan as unknown so the server can populate it.
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue("{}"),
		PlanValue:   types.StringValue("{}"),
		StateValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to be unknown for empty object config, got %q", resp.PlanValue)
	}
}

func TestUseServerValueIfEmpty_EmptyObjectWithWhitespace(t *testing.T) {
	t.Parallel()

	// Normalized JSON may have whitespace. An empty object should still trigger unknown.
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue("{ }"),
		PlanValue:   types.StringValue("{ }"),
		StateValue:  types.StringValue(`{"type":"object"}`),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if !resp.PlanValue.IsUnknown() {
		t.Errorf("expected plan value to be unknown for whitespace-only object config, got %q", resp.PlanValue)
	}
}

func TestUseServerValueIfEmpty_NonEmptyObject(t *testing.T) {
	t.Parallel()

	// User explicitly set a schema — keep it.
	schema := `{"type":"object","properties":{"name":{"type":"string"}}}`
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue(schema),
		PlanValue:   types.StringValue(schema),
		StateValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != schema {
		t.Errorf("expected plan value to remain %q, got %q", schema, resp.PlanValue.ValueString())
	}
}

func TestUseServerValueIfEmpty_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Invalid JSON — keep the config value and let validation catch it.
	invalid := "not-json"
	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue(invalid),
		PlanValue:   types.StringValue(invalid),
		StateValue:  types.StringNull(),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	resources.UseServerValueIfEmpty().PlanModifyString(context.Background(), req, resp)

	if resp.PlanValue.ValueString() != invalid {
		t.Errorf("expected plan value to remain %q for invalid JSON, got %q", invalid, resp.PlanValue.ValueString())
	}
}
