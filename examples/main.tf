terraform {
  required_providers {
    prefect = {
      source  = "prefecthq/prefect"
      version = "0.1"
    }
  }
}

provider "prefect" {
}
