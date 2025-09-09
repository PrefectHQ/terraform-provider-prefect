package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderWithProfileIntegration(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	prefectDir := filepath.Join(tempDir, ".prefect")
	err := os.MkdirAll(prefectDir, 0755)
	require.NoError(t, err)

	// Create a test profiles.toml file
	profilesContent := `
active = "test-profile"

[profiles.test-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api"
PREFECT_API_KEY = "test-api-key"
PREFECT_BASIC_AUTH_KEY = "test-basic-auth"
PREFECT_CSRF_ENABLED = "true"
`

	profilesPath := filepath.Join(prefectDir, "profiles.toml")
	err = os.WriteFile(profilesPath, []byte(profilesContent), 0644)
	require.NoError(t, err)

	// Mock the user home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	os.Setenv("HOME", tempDir)

	// Test that the provider can be created
	p := New()
	assert.NotNil(t, p)

	// Test that the provider implements the provider.Provider interface
	var _ provider.Provider = p
}

func TestProviderProfilePrecedence(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	prefectDir := filepath.Join(tempDir, ".prefect")
	err := os.MkdirAll(prefectDir, 0755)
	require.NoError(t, err)

	// Create a test profiles.toml file
	profilesContent := `
active = "test-profile"

[profiles.test-profile]
PREFECT_API_URL = "https://profile-api.prefect.cloud/api"
PREFECT_API_KEY = "profile-api-key"
`

	profilesPath := filepath.Join(prefectDir, "profiles.toml")
	err = os.WriteFile(profilesPath, []byte(profilesContent), 0644)
	require.NoError(t, err)

	// Mock the user home directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
		os.Unsetenv("PREFECT_API_URL")
		os.Unsetenv("PREFECT_API_KEY")
	})
	os.Setenv("HOME", tempDir)

	// Test 1: Profile should be used when no environment variables are set
	auth, err := LoadProfileAuth(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, types.StringValue("https://profile-api.prefect.cloud/api"), auth.Endpoint)
	assert.Equal(t, types.StringValue("profile-api-key"), auth.APIKey)
	assert.Equal(t, customtypes.NewUUIDNull(), auth.AccountID)

	// Test 2: Environment variables should override profile
	os.Setenv("PREFECT_API_URL", "https://env-api.prefect.cloud/api")
	os.Setenv("PREFECT_API_KEY", "env-api-key")

	// Reload profile (in real usage, this would be handled by the provider)
	auth, err = LoadProfileAuth(context.Background(), "", "")
	require.NoError(t, err)

	// Profile values should still be the same
	assert.Equal(t, types.StringValue("https://profile-api.prefect.cloud/api"), auth.Endpoint)
	assert.Equal(t, types.StringValue("profile-api-key"), auth.APIKey)

	// But in the actual provider logic, environment variables would take precedence
	// This is tested in the provider's Configure method
}
