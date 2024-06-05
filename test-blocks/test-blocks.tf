terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.32.0"
    }
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

provider "prefect" {
  # endpoint = "http://localhost:4200"
  # account_id = "9b649228-0419-40e1-9e0d-44954b5c0ab6"
  # api_key    = "pnu_XkEJQGRfJvOX0173wTvHwn39Y5VSJb08O7R4"
  workspace_id = "45cfa7c6-e136-471c-859b-3be89d0a99ce"
}


provider "google" {
  project = "prefect-sbx-eddiepark"
  region  = "us-east1"
}

# resource "prefect_block" "foo" {
#   name      = "foo"
#   type_slug = "secret"
#   data  = jsonencode({
#     "value" : "bar"
#   })
# }

resource "prefect_block" "aws_credentials_from_file" {
  name = "production-aws"

  # prefect block type ls
  type_slug = "aws-credentials"

  # prefect block type inspect aws-credentials
  data = file("./aws-credentials.json")
}


resource "google_service_account" "test_bot" {
  display_name = "test-bot"
  account_id   = "test-bot"
}

resource "google_service_account_key" "test_bot" {
  service_account_id = google_service_account.test_bot.id
}

resource "prefect_block" "gcp_credentials_key" {
  name = "staging-gcp"

  # prefect block type ls
  type_slug = "gcp-credentials"

  # prefect block type inspect gcp-credentials
  data = jsonencode({
    "project" : "prefect-sbx-eddiepark",
    "service_account_info" : base64decode(google_service_account_key.test_bot.private_key)
  })
}
