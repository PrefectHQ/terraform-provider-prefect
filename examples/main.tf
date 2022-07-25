terraform {
  required_providers {
    prefect = {
      source = "egistry.terraform.io/PrefectHQ/prefect"
    }
  }
  required_version = ">= 1.0"
}

provider "prefect" {

}

resource "prefect_project" "test" {
  name        = "test"
  description = "my test project"
}

resource "prefect_service_account" "test" {
  name = "test"
  role = "USER"
  api_keys = [
    {
      name       = "key1"
      expiration = "2015-10-21T00:00:00+11:00"
    },
    {
      name       = "key2"
      expiration = "2015-10-21T00:00:00+11:00"
    }
  ]
}

resource "aws_secretsmanager_secret" "this" {
  name = "prefect-svc.apikey"
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id     = aws_secretsmanager_secret.this.id
  secret_string = prefect_service_account.test.api_keys[0].key
}
