package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/prefecthq/terraform-provider-prefect/internal/provider"
)

func envFromOS() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}
	return "HOME"
}

func TestLoadPrefectProviderModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		profilesContent string
		expectedAuth    *provider.PrefectProviderModel
		expectError     bool
	}{
		{
			name: "valid profile with all settings",
			profilesContent: `
active = "test-profile"

[profiles.test-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"
PREFECT_API_KEY = "test-api-key"
PREFECT_API_AUTH_STRING = "test-basic-auth"
PREFECT_CSRF_ENABLED = "true"
`,
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint:     types.StringValue("https://api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"),
				APIKey:       types.StringValue("test-api-key"),
				BasicAuthKey: types.StringValue("test-basic-auth"),
				CSRFEnabled:  types.BoolValue(true),
			},
			expectError: false,
		},
		{
			name: "valid profile with partial settings",
			profilesContent: `
active = "minimal-profile"

[profiles.minimal-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api"
PREFECT_API_KEY = "minimal-key"
`,
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint: types.StringValue("https://api.prefect.cloud/api"),
				APIKey:   types.StringValue("minimal-key"),
			},
			expectError: false,
		},
		{
			name: "no active profile",
			profilesContent: `
[profiles.test-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api"
`,
			expectedAuth: &provider.PrefectProviderModel{},
			expectError:  false,
		},
		{
			name: "active profile not found",
			profilesContent: `
active = "nonexistent-profile"

[profiles.test-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api"
`,
			expectedAuth: nil,
			expectError:  true,
		},
		{
			name: "invalid TOML",
			profilesContent: `
active = "test-profile"
invalid toml content
`,
			expectedAuth: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the test
			tempDir := t.TempDir()
			prefectDir := filepath.Join(tempDir, ".prefect")
			err := os.MkdirAll(prefectDir, 0o755)
			require.NoError(t, err)

			// Write the profiles content to a temporary file
			profilesPath := filepath.Join(prefectDir, "profiles.toml")
			err = os.WriteFile(profilesPath, []byte(tt.profilesContent), 0o600)
			require.NoError(t, err)

			// Mock the user home directory
			env := envFromOS()
			originalHome := os.Getenv(env)
			t.Cleanup(func() {
				os.Setenv(env, originalHome)
			})
			os.Setenv(env, tempDir)

			// Test the function
			auth, err := provider.LoadProfileAuth(context.Background(), "", "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuth, auth)
			}
		})
	}
}

func TestLoadPrefectProviderModel_NoProfilesFile(t *testing.T) {
	t.Parallel()

	// Create a temporary directory without a profiles file
	tempDir := t.TempDir()

	// Mock the user home directory
	env := envFromOS()
	originalHome := os.Getenv(env)
	t.Cleanup(func() {
		os.Setenv(env, originalHome)
	})
	os.Setenv(env, tempDir)

	// Test the function
	auth, err := provider.LoadProfileAuth(context.Background(), "", "")

	assert.NoError(t, err)
	assert.Equal(t, &provider.PrefectProviderModel{}, auth)
}

func TestLoadPrefectProviderModel_EmptyProfilesFile(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with an empty profiles file
	tempDir := t.TempDir()
	prefectDir := filepath.Join(tempDir, ".prefect")
	err := os.MkdirAll(prefectDir, 0o755)
	require.NoError(t, err)

	profilesPath := filepath.Join(prefectDir, "profiles.toml")
	err = os.WriteFile(profilesPath, []byte(""), 0o600)
	require.NoError(t, err)

	// Mock the user home directory
	env := envFromOS()
	originalHome := os.Getenv(env)
	t.Cleanup(func() {
		os.Setenv(env, originalHome)
	})
	os.Setenv(env, tempDir)

	// Test the function
	auth, err := provider.LoadProfileAuth(context.Background(), "", "")

	assert.NoError(t, err)
	assert.Equal(t, &provider.PrefectProviderModel{}, auth)
}

func TestLoadPrefectProviderModel_SpecificProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		profilesContent string
		profileName     string
		expectedAuth    *provider.PrefectProviderModel
		expectError     bool
	}{
		{
			name: "load specific profile that exists",
			profilesContent: `
active = "default-profile"

[profiles.default-profile]
PREFECT_API_URL = "https://default-api.prefect.cloud/api"
PREFECT_API_KEY = "default-key"

[profiles.prod-profile]
PREFECT_API_URL = "https://api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"
PREFECT_API_KEY = "prod-key"
`,
			profileName: "prod-profile",
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint: types.StringValue("https://api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"),
				APIKey:   types.StringValue("prod-key"),
			},
			expectError: false,
		},
		{
			name: "load specific profile that doesn't exist",
			profilesContent: `
active = "default-profile"

[profiles.default-profile]
PREFECT_API_URL = "https://default-api.prefect.cloud/api"
`,
			profileName:  "nonexistent-profile",
			expectedAuth: nil,
			expectError:  true,
		},
		{
			name: "empty profile name uses active profile",
			profilesContent: `
active = "active-profile"

[profiles.active-profile]
PREFECT_API_URL = "https://active-api.prefect.cloud/api"
PREFECT_API_KEY = "active-key"
`,
			profileName: "",
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint: types.StringValue("https://active-api.prefect.cloud/api"),
				APIKey:   types.StringValue("active-key"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the test
			tempDir := t.TempDir()
			prefectDir := filepath.Join(tempDir, ".prefect")
			err := os.MkdirAll(prefectDir, 0o755)
			require.NoError(t, err)

			// Write the profiles content to a temporary file
			profilesPath := filepath.Join(prefectDir, "profiles.toml")
			err = os.WriteFile(profilesPath, []byte(tt.profilesContent), 0o600)
			require.NoError(t, err)

			// Mock the user home directory
			env := envFromOS()
			originalHome := os.Getenv(env)
			t.Cleanup(func() {
				os.Setenv(env, originalHome)
			})
			os.Setenv(env, tempDir)

			// Test the function
			auth, err := provider.LoadProfileAuth(context.Background(), tt.profileName, "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuth, auth)
			}
		})
	}
}

func TestLoadPrefectProviderModel_CustomProfileFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		profilesContent string
		profileName     string
		profileFilePath string
		expectedAuth    *provider.PrefectProviderModel
		expectError     bool
	}{
		{
			name: "load from custom profile file path",
			profilesContent: `
active = "custom-profile"

[profiles.custom-profile]
PREFECT_API_URL = "https://custom-api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"
PREFECT_API_KEY = "custom-key"
`,
			profileName:     "",
			profileFilePath: "/tmp/custom-profiles.toml",
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint: types.StringValue("https://custom-api.prefect.cloud/api/accounts/123e4567-e89b-12d3-a456-426614174000/workspaces/987fcdeb-51a2-43d1-b123-456789abcdef"),
				APIKey:   types.StringValue("custom-key"),
			},
			expectError: false,
		},
		{
			name: "load specific profile from custom file",
			profilesContent: `
active = "default-profile"

[profiles.default-profile]
PREFECT_API_URL = "https://default-api.prefect.cloud/api"

[profiles.custom-profile]
PREFECT_API_URL = "https://custom-api.prefect.cloud/api"
PREFECT_API_KEY = "custom-key"
`,
			profileName:     "custom-profile",
			profileFilePath: "/tmp/custom-profiles.toml",
			expectedAuth: &provider.PrefectProviderModel{
				Endpoint: types.StringValue("https://custom-api.prefect.cloud/api"),
				APIKey:   types.StringValue("custom-key"),
			},
			expectError: false,
		},
		{
			name:            "custom profile file does not exist",
			profilesContent: "",
			profileName:     "",
			profileFilePath: "/nonexistent/path/profiles.toml",
			expectedAuth:    &provider.PrefectProviderModel{},
			expectError:     false,
		},
		{
			name: "custom profile file with invalid TOML",
			profilesContent: `
active = "test-profile"
invalid toml content
`,
			profileName:     "",
			profileFilePath: "/tmp/invalid-profiles.toml",
			expectedAuth:    nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the test
			tempDir := t.TempDir()

			// Set up the custom profile file path
			var customProfilePath string
			if tt.profileFilePath != "" {
				if tt.profileFilePath == "/tmp/custom-profiles.toml" || tt.profileFilePath == "/tmp/invalid-profiles.toml" {
					customProfilePath = filepath.Join(tempDir, "custom-profiles.toml")
				} else {
					customProfilePath = tt.profileFilePath
				}
			}

			// Write the profiles content to the custom file if it exists
			if tt.profilesContent != "" && customProfilePath != "" {
				err := os.WriteFile(customProfilePath, []byte(tt.profilesContent), 0o600)
				require.NoError(t, err)
			}

			// Test the function
			auth, err := provider.LoadProfileAuth(context.Background(), tt.profileName, customProfilePath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuth, auth)
			}
		})
	}
}
