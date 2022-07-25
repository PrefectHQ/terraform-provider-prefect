# Configure the Prefect provider using the required_providers stanza.
# You may optionally use a version directive to prevent breaking
# changes occurring unannounced.
terraform {
  required_providers {
    prefect = {
      source = "PrefectHQ/prefect"
    }
  }
  required_version = ">= 1.0"
}

provider "prefect" {
  # Or omit this for api_server to be read from the
  # PREFECT__CLOUD__API environment variable.
  # If neither are set will default to https://api.prefect.io
  api_server = var.api_server

  # Or omit this for api_key to be read from the
  # PREFECT__CLOUD__API_KEY environment variable.
  api_key = var.api_key
}
