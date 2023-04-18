terraform {
  required_providers {
    prefect = {
      version = "0.2"
      source  = "hashicorp.com/edu/prefect"
    }
  }
}

provider "prefect" {}