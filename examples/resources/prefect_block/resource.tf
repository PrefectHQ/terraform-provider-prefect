# example:
# the data property will hold json encoded data
resource "prefect_block" "secret" {
  name      = "foo"
  type_slug = "secret"

  data = jsonencode({
    "value" = "bar"
  })

  # set the workpace_id attribute on the provider OR the resource
  workspace_id = "<workspace UUID>"
}
# example:
# you can also use a write-only attribute for the data field
resource "prefect_block" "secret_write_only" {
  name = "foo-write-only"

  # prefect block type ls
  type_slug = "aws-credentials"

  # prefect block type inspect aws-credentials
  data_wo = jsonencode({
    "value" = "bar"
  })

  # provide the version to control when to update the block data
  data_wo_version = 1
}

# example:
# you can also use file() to import a JSON file
# or even a templatefile() to import a JSON template file
resource "prefect_block" "aws_credentials_from_file" {
  name = "production-aws"

  # prefect block type ls
  type_slug = "aws-credentials"

  # prefect block type inspect aws-credentials
  data = file("./aws-credentials.json")
}

# example:
# Here's an end-to-end example of the gcp-credentials block
# holding a GCP service account key
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
    "project"              = "my-gcp-project",
    "service_account_info" = jsondecode(base64decode(google_service_account_key.test_bot.private_key))
  })
}

# example:
# some resources need to be referenced using a "$ref" key
resource "prefect_block" "my_dbt_cli_profile" {
  name      = "my-dbt-cli-profile"
  type_slug = "dbt-cli-profile"

  data = jsonencode({
    "name"   = "mine"
    "target" = "prefect-dbt-profile"
  })
}
resource "prefect_block" "my_dbt_run_operation_block" {
  name      = "my-dbt-operations"
  type_slug = "dbt-core-operation"

  # note the "$ref" key wrapping the "block_document_id" reference
  data = jsonencode({
    "commands"        = ["dbt deps", "dbt seed", "dbt run"]
    "dbt_cli_profile" = { "$ref" : { "block_document_id" : prefect_block.my_dbt_cli_profile.id } }
  })
}
