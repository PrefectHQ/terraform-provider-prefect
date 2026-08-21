package resources

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

func TestCopyDeploymentToModelStorageDocumentID(t *testing.T) {
	t.Parallel()

	storageDocumentID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tests := []struct {
		name              string
		storageDocumentID uuid.UUID
		want              customtypes.UUIDValue
	}{
		{
			name:              "absent",
			storageDocumentID: uuid.Nil,
			want:              customtypes.NewUUIDNull(),
		},
		{
			name:              "present",
			storageDocumentID: storageDocumentID,
			want:              customtypes.NewUUIDValue(storageDocumentID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var model DeploymentResourceModel
			diags := CopyDeploymentToModel(context.Background(), &api.Deployment{
				StorageDocumentID: tt.storageDocumentID,
			}, &model)
			if diags.HasError() {
				t.Fatalf("CopyDeploymentToModel returned errors: %v", diags)
			}

			if !model.StorageDocumentID.Equal(tt.want) {
				t.Errorf("storage_document_id = %s, want %s", model.StorageDocumentID.ValueString(), tt.want.ValueString())
			}
		})
	}
}
