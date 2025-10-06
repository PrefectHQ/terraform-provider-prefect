package planmodifiers

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

// normalizeOpenAPISchema is a plan modifier that normalizes OpenAPI schemas.
type normalizeOpenAPISchema struct{}

// NormalizeOpenAPISchema returns a plan modifier that normalizes OpenAPI schemas
// to the format expected by Prefect OSS/Cloud. This ensures consistency between
// planned and actual values, preventing "inconsistent result after apply" errors.
//
//nolint:ireturn // required by Terraform API
func NormalizeOpenAPISchema() planmodifier.String {
	return normalizeOpenAPISchema{}
}

// Description returns a human-readable description of the plan modifier.
func (m normalizeOpenAPISchema) Description(_ context.Context) string {
	return "Normalizes empty or partial OpenAPI schemas to the canonical format: {\"type\": \"object\", \"properties\": {}}"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m normalizeOpenAPISchema) MarkdownDescription(_ context.Context) string {
	return "Normalizes empty or partial OpenAPI schemas to the canonical format: `{\"type\": \"object\", \"properties\": {}}`"
}

// PlanModifyString implements the plan modification logic.
func (m normalizeOpenAPISchema) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the value is null or unknown, don't modify it
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// Get the planned value as a jsontypes.Normalized value
	var plannedValue jsontypes.Normalized
	diags := req.Plan.GetAttribute(ctx, req.Path, &plannedValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the value is null, don't modify it
	if plannedValue.IsNull() {
		return
	}

	// Unmarshal the JSON to a map
	var schemaMap map[string]interface{}
	diags = plannedValue.Unmarshal(&schemaMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normalize the schema
	normalizedSchema := helpers.NormalizeParameterOpenAPISchema(schemaMap)

	// Marshal back to JSON
	normalizedJSON, err := json.Marshal(normalizedSchema)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error normalizing parameter_openapi_schema",
			"Could not marshal normalized schema: "+err.Error(),
		)

		return
	}

	// Set the normalized value back in the plan as a types.String
	resp.PlanValue = types.StringValue(string(normalizedJSON))
}
