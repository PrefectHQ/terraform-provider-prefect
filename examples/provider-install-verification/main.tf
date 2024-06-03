terraform {
  required_providers {
    prefect = {
      source = "registry.terraform.io/prefecthq/prefect"
    }
  }
}

provider "prefect" {
  endpoint = "http://localhost:4200/api"
}

resource "prefect_work_pool" "example" {
  name        = "my-work-pool"
  type        = "kubernetes"
  description = "example work pool"
  paused      = true
}
