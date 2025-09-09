terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

# Authentication configuration precedence:
# 1. Provider block attributes (highest priority)
# 2. Environment variables
# 3. Prefect profile file (lowest priority)
#
# The provider will attempt to automatically load authentication from your Prefect profile
# located at ~/.prefect/profiles.toml if no explicit configuration is provided.

# By default, the provider points to Prefect Cloud
# and you can pass in your API key and account ID
# via variables or static inputs.
provider "prefect" {
  api_key    = var.prefect_api_key
  account_id = var.prefect_account_id
}

# You can also pass in your API key and account ID
# implicitly via environment variables, such as
# PREFECT_API_KEY and PREFECT_CLOUD_ACCOUNT_ID.
provider "prefect" {}

# You also have the option to link the provider instance
# to your specific workspace, if this fits your use case.
provider "prefect" {
  api_key      = var.prefect_api_key
  account_id   = var.prefect_account_id
  workspace_id = var.prefect_workspace_id
}

# Use a specific profile instead of the active one
provider "prefect" {
  profile = "prod-profile"
  # All authentication will be loaded from the "prod-profile"
}

# Use a custom profile file location
provider "prefect" {
  profile_file = "/path/to/custom/profiles.toml"
  # All authentication will be loaded from the custom profiles file
}

# Use a specific profile from a custom file
provider "prefect" {
  profile      = "dev-profile"
  profile_file = "/path/to/dev/profiles.toml"
  # Load "dev-profile" from the custom profiles file
}

# You can still override specific settings from the profile
provider "prefect" {
  profile = "my-profile"
  # This will override the API URL from the profile
  endpoint = "https://custom-api.prefect.cloud/api"

  # Other settings will still be loaded from the profile
}

# You also have the option to specify the account and workspace
# in the `endpoint` attribute. This is the same format used for
# the `PREFECT_API_KEY` value used in the Prefect CLI configuration file.
provider "prefect" {
  endpoint = "https://api.prefect.cloud/api/accounts/<account_id>/workspaces/<workspace_id>"
}

# Finally, in rare occasions, you also have the option
# to point the provider to a locally running Prefect Server,
# with a limited set of functionality from the provider.
provider "prefect" {
  endpoint = "http://localhost:4200"
}
