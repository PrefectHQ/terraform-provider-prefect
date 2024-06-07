terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

data "prefect_account_members" "all_members" {}
