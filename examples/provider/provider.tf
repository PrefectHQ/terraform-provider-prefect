terraform {
  required_providers {
    prefect = {
      source = "hashicorp.com/prefecthq/prefect"
    }
  }
}

provider "prefect" {
  endpoint   = "https://api.prefect.cloud"
  api_key    = var.prefect_api_key
  account_id = var.prefect_account_id
}
