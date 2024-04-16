terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

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

# Finally, in rare occasions, you also have the option
# to point the provider to a locally running Prefect Server,
# with a limited set of functionality from the provider.
provider "prefect" {
  endpoint = "http://localhost:4200"
}
