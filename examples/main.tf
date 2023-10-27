terraform {
  required_providers {
    prefect = {
      source  = "prefecthq/prefect"
      version = "0.1"
    }
  }
}

provider "prefect" {
  api_key    = var.PREFECT_API_KEY
  account_id = var.PREFECT_ACCOUNT_ID
}
