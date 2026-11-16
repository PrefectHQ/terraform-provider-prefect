package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// fakeBlockSchemasClient returns readErr from every BlockSchemas().Read call.
type fakeBlockSchemasClient struct {
	api.PrefectClient
	api.BlockSchemaClient

	readErr error
}

//nolint:ireturn // required by api.PrefectClient
func (f *fakeBlockSchemasClient) BlockSchemas(_ uuid.UUID, _ uuid.UUID) (api.BlockSchemaClient, error) {
	return f, nil
}

func (f *fakeBlockSchemasClient) Read(_ context.Context, _ uuid.UUID) (*api.BlockSchema, error) {
	return nil, f.readErr
}

func newBlockSchemaReadTest(t *testing.T, readErr error) (*BlockSchemaResource, tfsdk.State) {
	t.Helper()

	r := &BlockSchemaResource{client: &fakeBlockSchemasClient{readErr: readErr}}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "schema: %v", schemaResp.Diagnostics)

	model := BlockSchemaResourceModel{
		BaseModel: BaseModel{
			ID:      customtypes.NewUUIDValue(uuid.New()),
			Created: customtypes.NewTimestampNull(),
			Updated: customtypes.NewTimestampNull(),
		},
		AccountID:    customtypes.NewUUIDNull(),
		WorkspaceID:  customtypes.NewUUIDNull(),
		Checksum:     types.StringValue("sha256:abc"),
		Fields:       jsontypes.NewNormalizedValue(`{}`),
		BlockTypeID:  customtypes.NewUUIDValue(uuid.New()),
		BlockType:    types.StringValue("my-block"),
		Capabilities: types.ListNull(types.StringType),
		Version:      types.StringValue("1"),
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	diags := state.Set(context.Background(), &model)
	require.False(t, diags.HasError(), "state: %v", diags)

	return r, state
}

func TestBlockSchemaReadNotFoundRemovesResource(t *testing.T) {
	t.Parallel()

	r, state := newBlockSchemaReadTest(t, errors.New("failed to get block schema: status_code=404, error=<nil>, body={\"detail\":\"Block schema not found\"}"))

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	require.False(t, resp.Diagnostics.HasError(), "diags: %v", resp.Diagnostics)
	assert.True(t, resp.State.Raw.IsNull(), "expected the block schema to be removed from state")
}

func TestBlockSchemaReadErrorReturnsDiagnostic(t *testing.T) {
	t.Parallel()

	r, state := newBlockSchemaReadTest(t, errors.New("failed to get block schema: status_code=500, error=<nil>, body=internal error"))

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	assert.True(t, resp.Diagnostics.HasError())
}
