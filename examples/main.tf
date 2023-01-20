terraform {
  required_providers {
    prefect = {
      source = "PrefectHQ/prefect"
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
      expiration = "2015-10-21T00:00:00+00:00"
    },
    {
      name       = "key2"
      expiration = "2015-10-21T00:00:00+00:00"
    }
  ]
}

# example showing how to store the api key in an AWS Secrets Manager secret

# resource "aws_secretsmanager_secret" "this" {
#   name = "prefect-svc.apikey"
# }

# resource "aws_secretsmanager_secret_version" "this" {
#   secret_id     = aws_secretsmanager_secret.this.id
#   secret_string = prefect_service_account.test.api_keys[0].key
# }
