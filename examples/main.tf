terraform {
  required_providers {
    prefect = {
      source  = "prefecthq/prefect"
      version = "0.1"
    }
    time = {
      source = "hashicorp/time"
      version = "0.9.1"
    }
  }
}

provider "prefect" {
}

provider "time" {
}
