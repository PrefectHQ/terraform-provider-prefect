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
