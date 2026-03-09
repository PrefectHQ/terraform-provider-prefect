package resources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseServerValueIfEmpty returns a plan modifier that marks the planned value
// as unknown when the user's config value is an empty JSON object ("{}").
// This allows the server to populate the field without causing a taint loop.
//
// When the config is null (not set by the user), the modifier preserves
// the prior state value on updates, similar to UseStateForUnknown.
//
//nolint:ireturn // standard pattern for terraform plan modifiers
func UseServerValueIfEmpty() planmodifier.String {
	return useServerValueIfEmptyModifier{}
}

type useServerValueIfEmptyModifier struct{}

func (m useServerValueIfEmptyModifier) Description(_ context.Context) string {
	return "Accepts the server's value when the config is an empty JSON object or null."
}

func (m useServerValueIfEmptyModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useServerValueIfEmptyModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the config value is unknown, let the framework handle it.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// If the config value is null (user didn't set it):
	// - On create, leave as unknown so the server can populate it.
	// - On update, preserve the prior state value.
	if req.ConfigValue.IsNull() {
		if !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}

		return
	}

	// If the config value is a non-empty JSON object, keep it as-is.
	// An empty object means "let the server decide."
	configStr := req.ConfigValue.ValueString()

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &obj); err != nil {
		// Not valid JSON; keep config value and let validation catch it.
		return
	}

	if len(obj) == 0 {
		resp.PlanValue = types.StringUnknown()

		return
	}

	// Non-empty JSON: the user intentionally set a schema, keep it.
}
