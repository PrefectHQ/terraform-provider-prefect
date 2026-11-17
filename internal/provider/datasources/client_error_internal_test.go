package datasources

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

// failingClient fails to build the account and deployment clients, like the
// real client does when no account_id (or, for Prefect Cloud, no
// workspace_id) is set on the provider or the data source.
type failingClient struct {
	api.PrefectClient
}

//nolint:ireturn // required by api.PrefectClient
func (failingClient) Accounts(_ uuid.UUID) (api.AccountsClient, error) {
	return nil, errors.New("accountID must be set")
}

//nolint:ireturn // required by api.PrefectClient
func (failingClient) Deployments(_, _ uuid.UUID) (api.DeploymentsClient, error) {
	return nil, errors.New("prefect Cloud endpoints require an account_id and workspace_id")
}

func readWithClientError(t *testing.T, ds datasource.DataSource, attrs map[string]tftypes.Value) *datasource.ReadResponse {
	t.Helper()

	var schemaResp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "schema: %v", schemaResp.Diagnostics)

	objType, ok := schemaResp.Schema.Type().TerraformType(context.Background()).(tftypes.Object)
	require.True(t, ok)

	values := map[string]tftypes.Value{}
	for name, typ := range objType.AttributeTypes {
		if v, set := attrs[name]; set {
			values[name] = v
		} else {
			values[name] = tftypes.NewValue(typ, nil)
		}
	}

	config := tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objType, values)}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}

	assert.NotPanics(t, func() {
		ds.Read(context.Background(), datasource.ReadRequest{Config: config}, resp)
	})

	return resp
}

func TestAccountReadClientErrorReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	resp := readWithClientError(t, &AccountDataSource{client: failingClient{}}, nil)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestDeploymentReadClientErrorReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	resp := readWithClientError(t, &deploymentDataSource{client: failingClient{}}, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, uuid.NewString()),
	})

	assert.True(t, resp.Diagnostics.HasError())
}
