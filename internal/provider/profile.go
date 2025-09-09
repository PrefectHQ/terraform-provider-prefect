package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProfileConfig represents the configuration for a single Prefect profile
type ProfileConfig map[string]string

// ProfilesConfig represents the entire profiles.toml file structure
type ProfilesConfig struct {
	ActiveProfile string                   `toml:"active"`
	Profiles      map[string]ProfileConfig `toml:"profiles"`
}

// LoadProfileAuth loads authentication information from the specified Prefect profile.
// If profileName is empty, uses the active profile from the profiles.toml file.
// If profileFilePath is empty, uses the default location ~/.prefect/profiles.toml.
func LoadProfileAuth(_ context.Context, profileName string, profileFilePath string) (*PrefectProviderModel, error) {
	// Determine the profiles file path
	var profilesPath string
	if profileFilePath != "" {
		profilesPath = profileFilePath
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		profilesPath = filepath.Join(homeDir, ".prefect", "profiles.toml")
	}

	// Check if the profiles file exists
	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		// Profiles file doesn't exist, return empty auth
		return &PrefectProviderModel{}, nil
	}

	// Read and parse the profiles file
	var config ProfilesConfig
	if _, err := toml.DecodeFile(profilesPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse profiles.toml: %w", err)
	}

	// Determine which profile to use
	var targetProfileName string
	if profileName != "" {
		targetProfileName = profileName
	} else if config.ActiveProfile != "" {
		targetProfileName = config.ActiveProfile
	} else {
		return &PrefectProviderModel{}, nil
	}

	// Get the target profile
	targetProfile, exists := config.Profiles[targetProfileName]
	if !exists {
		if profileName != "" {
			return nil, fmt.Errorf("profile '%s' not found in profiles.toml", profileName)
		}
		return nil, fmt.Errorf("active profile '%s' not found in profiles.toml", config.ActiveProfile)
	}

	// Extract authentication settings from the profile
	auth := &PrefectProviderModel{}

	if apiURL, exists := targetProfile[envAPIURL]; exists {
		auth.Endpoint = types.StringValue(apiURL)
	}

	if apiKey, exists := targetProfile[envAPIKey]; exists {
		auth.APIKey = types.StringValue(apiKey)
	}

	if basicAuthKey, exists := targetProfile[envBasicAuthKey]; exists {
		auth.BasicAuthKey = types.StringValue(basicAuthKey)
	}

	// TODO: Deprecate PREFECT_BASIC_AUTH_KEY in favor of PREFECT_API_AUTH_STRING
	if basicAuthKey, exists := targetProfile["PREFECT_API_AUTH_STRING"]; exists {
		auth.BasicAuthKey = types.StringValue(basicAuthKey)
	}

	if csrfEnabled, exists := targetProfile[envCSRFEnabled]; exists {
		auth.CSRFEnabled = types.BoolValue(csrfEnabled == "true")
	}

	return auth, nil
}
