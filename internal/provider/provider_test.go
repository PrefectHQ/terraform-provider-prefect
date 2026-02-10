package provider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	provider "github.com/prefecthq/terraform-provider-prefect/internal/provider"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

func setStringAttr(attrs map[string]tftypes.Value, key string, value types.String) {
	if !value.IsNull() {
		attrs[key] = tftypes.NewValue(tftypes.String, value.ValueString())
	} else {
		attrs[key] = tftypes.NewValue(tftypes.String, nil)
	}
}

func setBoolAttr(attrs map[string]tftypes.Value, key string, value types.Bool) {
	if !value.IsNull() {
		attrs[key] = tftypes.NewValue(tftypes.Bool, value.ValueBool())
	} else {
		attrs[key] = tftypes.NewValue(tftypes.Bool, nil)
	}
}

func setUUIDAttr(attrs map[string]tftypes.Value, key string, value customtypes.UUIDValue) {
	if !value.IsNull() {
		attrs[key] = tftypes.NewValue(tftypes.String, value.ValueString())
	} else {
		attrs[key] = tftypes.NewValue(tftypes.String, nil)
	}
}

func newTestConfigureRequest(t *testing.T, model *provider.PrefectProviderModel) tfprovider.ConfigureRequest {
	t.Helper()

	// Manually construct tftypes.Object from PrefectProviderModel
	attrTypes := map[string]tftypes.Type{
		"endpoint":       tftypes.String,
		"api_key":        tftypes.String,
		"basic_auth_key": tftypes.String,
		"csrf_enabled":   tftypes.Bool,
		"custom_headers": tftypes.String,
		"account_id":     tftypes.String, // customtypes.UUIDType is based on string
		"workspace_id":   tftypes.String,
		"profile":        tftypes.String,
		"profile_file":   tftypes.String,
	}

	attrs := make(map[string]tftypes.Value)

	if model != nil {
		setStringAttr(attrs, "endpoint", model.Endpoint)
		setStringAttr(attrs, "api_key", model.APIKey)
		setStringAttr(attrs, "basic_auth_key", model.BasicAuthKey)
		setBoolAttr(attrs, "csrf_enabled", model.CSRFEnabled)
		setStringAttr(attrs, "custom_headers", model.CustomHeaders)
		setUUIDAttr(attrs, "account_id", model.AccountID)
		setUUIDAttr(attrs, "workspace_id", model.WorkspaceID)
		setStringAttr(attrs, "profile", model.Profile)
		setStringAttr(attrs, "profile_file", model.ProfileFile)
	}

	rawObject := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, attrs)

	p := &provider.PrefectProvider{}
	schemaResp := &tfprovider.SchemaResponse{}
	p.Schema(context.Background(), tfprovider.SchemaRequest{}, schemaResp)

	return tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw:    rawObject,
			Schema: &schemaResp.Schema,
		},
		TerraformVersion:   "1.6.0",
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}
}

// TestConfigure_InvalidURL returns error.
func TestConfigure_InvalidURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prov := &provider.PrefectProvider{}
	resp := &tfprovider.ConfigureResponse{}

	config := &provider.PrefectProviderModel{
		Endpoint: types.StringValue("XTes^%:98"),
		// APIKey is missing (null by default)
	}

	req := newTestConfigureRequest(t, config)

	prov.Configure(ctx, req, resp)

	var found bool
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityError && strings.Contains(d.Summary(), "Invalid Prefect API Endpoint") {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("expected error for invalid URL")
	}
}

// TestConfigure_CustomHeaders tests that valid custom headers can be configured.
func TestConfigure_CustomHeaders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prov := &provider.PrefectProvider{}
	resp := &tfprovider.ConfigureResponse{}

	customHeaders := `{"X-Custom-Header": "test-value", "X-Another-Header": "another-value"}`

	config := &provider.PrefectProviderModel{
		Endpoint:      types.StringValue("https://api.example.com"),
		CustomHeaders: types.StringValue(customHeaders),
	}

	req := newTestConfigureRequest(t, config)
	prov.Configure(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestConfigure_CustomHeadersInvalidJSON tests that invalid JSON returns an error.
func TestConfigure_CustomHeadersInvalidJSON(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prov := &provider.PrefectProvider{}
	resp := &tfprovider.ConfigureResponse{}

	config := &provider.PrefectProviderModel{
		Endpoint:      types.StringValue("https://api.example.com"),
		CustomHeaders: types.StringValue("invalid json"),
	}

	req := newTestConfigureRequest(t, config)
	prov.Configure(ctx, req, resp)

	var found bool
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityError && strings.Contains(d.Summary(), "Invalid Custom Headers JSON") {
			found = true

			break
		}
	}

	if !found {
		t.Fatalf("expected error for invalid custom headers JSON")
	}
}

// TestConfigure_CustomHeadersWithProtectedHeaders tests that protected headers are filtered.
func TestConfigure_CustomHeadersWithProtectedHeaders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prov := &provider.PrefectProvider{}
	resp := &tfprovider.ConfigureResponse{}

	// Include both valid and protected headers
	customHeaders := `{"X-Custom-Header": "test-value", "User-Agent": "should-be-filtered", "Prefect-Csrf-Token": "should-be-filtered"}`

	config := &provider.PrefectProviderModel{
		Endpoint:      types.StringValue("https://api.example.com"),
		CustomHeaders: types.StringValue(customHeaders),
	}

	req := newTestConfigureRequest(t, config)
	prov.Configure(ctx, req, resp)

	// Should not error when protected headers are provided
	// They should just be filtered out with a warning
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// TestConfigure_CustomHeadersEmpty tests that empty custom headers don't cause errors.
func TestConfigure_CustomHeadersEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prov := &provider.PrefectProvider{}
	resp := &tfprovider.ConfigureResponse{}

	config := &provider.PrefectProviderModel{
		Endpoint:      types.StringValue("https://api.example.com"),
		CustomHeaders: types.StringValue("{}"),
	}

	req := newTestConfigureRequest(t, config)
	prov.Configure(ctx, req, resp)

	// Empty custom headers should not cause an error
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}
