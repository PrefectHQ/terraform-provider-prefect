# example:
# the data property will hold json encoded data
resource "prefect_block" "secret" {
  name      = "foo"
  type_slug = "secret"

  data = jsonencode({
    "value" : "bar"
  })

  # set the workpace_id attribute on the provider OR the resource
  workspace_id = "<workspace UUID>"
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
    "project" : "my-gcp-project",
    "service_account_info" : base64decode(google_service_account_key.test_bot.private_key)
  })
}
