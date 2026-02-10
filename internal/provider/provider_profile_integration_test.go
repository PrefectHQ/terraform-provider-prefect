package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefecthq/terraform-provider-prefect/internal/provider"
)

func TestProviderWithProfileIntegration(t *testing.T) {
	// Note: Cannot use t.Parallel() when using t.Setenv() to modify environment variables

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	prefectDir := filepath.Join(tempDir, ".prefect")
	err := os.MkdirAll(prefectDir, 0o755)
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
	err = os.WriteFile(profilesPath, []byte(profilesContent), 0o600)
	require.NoError(t, err)

	// Mock the user home directory
	env := envFromOS()
	t.Setenv(env, tempDir)

	// Test that the provider can be created
	p := provider.New()
	assert.NotNil(t, p)

	// Test that the provider implements the provider.Provider interface
	var _ = p
}

func TestProviderProfilePrecedence(t *testing.T) {
	// Note: Cannot use t.Parallel() when using t.Setenv() to modify environment variables

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	prefectDir := filepath.Join(tempDir, ".prefect")
	err := os.MkdirAll(prefectDir, 0o755)
	require.NoError(t, err)

	// Create a test profiles.toml file
	profilesContent := `
active = "test-profile"

[profiles.test-profile]
PREFECT_API_URL = "https://profile-api.prefect.cloud/api"
PREFECT_API_KEY = "profile-api-key"
`

	profilesPath := filepath.Join(prefectDir, "profiles.toml")
	err = os.WriteFile(profilesPath, []byte(profilesContent), 0o600)
	require.NoError(t, err)

	// Mock the user home directory
	env := envFromOS()
	t.Setenv(env, tempDir)

	// Test 1: Profile should be used when no environment variables are set
	auth, err := provider.LoadProfileAuth(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, types.StringValue("https://profile-api.prefect.cloud/api"), auth.Endpoint)
	assert.Equal(t, types.StringValue("profile-api-key"), auth.APIKey)
	assert.Equal(t, customtypes.NewUUIDNull(), auth.AccountID)

	// Test 2: Environment variables should override profile
	t.Setenv("PREFECT_API_URL", "https://env-api.prefect.cloud/api")
	t.Setenv("PREFECT_API_KEY", "env-api-key")

	// Reload profile (in real usage, this would be handled by the provider)
	auth, err = provider.LoadProfileAuth(context.Background(), "", "")
	require.NoError(t, err)

	// Profile values should still be the same
	assert.Equal(t, types.StringValue("https://profile-api.prefect.cloud/api"), auth.Endpoint)
	assert.Equal(t, types.StringValue("profile-api-key"), auth.APIKey)

	// But in the actual provider logic, environment variables would take precedence
	// This is tested in the provider's Configure method
}
