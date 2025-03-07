package provider_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestURLContainsIDs(t *testing.T) {
	t.Parallel()

	testAccountID := uuid.New()
	testWorkspaceID := uuid.New()
	testAPIURLWithIDs := fmt.Sprintf("https://api.prefect.cloud/api/accounts/%s/workspaces/%s", testAccountID, testWorkspaceID)
	testAPIURLWithAccountID := fmt.Sprintf("https://api.prefect.cloud/api/accounts/%s", testAccountID)

	tests := []struct {
		name   string
		apiURL string
		want   bool
	}{
		{
			name:   "URL without IDs",
			apiURL: "https://api.prefect.cloud",
			want:   false,
		},
		{
			name:   "URL with IDs",
			apiURL: testAPIURLWithIDs,
			want:   true,
		},
		{
			name:   "URL with account ID and no workspace ID",
			apiURL: testAPIURLWithAccountID,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := provider.URLContainsIDs(tt.apiURL)

			assert.Equal(t, tt.want, got, "Expected %+v but got %+v", tt.want, got)
		})
	}
}

func TestGetUUIDFromPath(t *testing.T) {
	t.Parallel()

	testAccountID := uuid.New()
	testWorkspaceID := uuid.New()
	pathWithIDs := fmt.Sprintf("/api/accounts/%s/workspaces/%s", testAccountID, testWorkspaceID)
	pathWithInvalidIDs := fmt.Sprintf("/api/accounts/%s/workspaces/%s", "invalid", "invalid")

	tests := []struct {
		name            string
		path            string
		wantAccountID   uuid.UUID
		wantWorkspaceID uuid.UUID
		wantErr         bool
	}{
		{
			name:            "Fail to get IDs when path does not contain them",
			path:            "https://api.prefect.cloud",
			wantAccountID:   uuid.Nil,
			wantWorkspaceID: uuid.Nil,
			wantErr:         true,
		},
		{
			name:            "Get IDs when path contains them",
			path:            pathWithIDs,
			wantAccountID:   testAccountID,
			wantWorkspaceID: testWorkspaceID,
			wantErr:         false,
		},
		{
			name:            "Fail when UUID is invalid",
			path:            pathWithInvalidIDs,
			wantAccountID:   uuid.Nil,
			wantWorkspaceID: uuid.Nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotAccountID, accountErr := provider.GetAccountIDFromPath(tt.path)
			if !tt.wantErr {
				assert.NoError(t, accountErr)
			}
			assert.Equal(t, tt.wantAccountID, gotAccountID, "Expected account ID %+v but got %+v", tt.wantAccountID, gotAccountID)

			gotWorkspaceID, accountErr := provider.GetWorkspaceIDFromPath(tt.path)
			if !tt.wantErr {
				assert.NoError(t, accountErr)
			}
			assert.Equal(t, tt.wantWorkspaceID, gotWorkspaceID, "Expected workspace ID %+v but got %+v", tt.wantWorkspaceID, gotWorkspaceID)
		})
	}
}
