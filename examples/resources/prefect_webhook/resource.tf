resource "prefect_webhook" "example" {
  name        = "my-webhook"
  description = "This is a webhook"
  enabled     = true
  template = jsonencode({
    event = "model.refreshed"
    resource = {
      "prefect.resource.id"   = "product.models.{{ body.model }}"
      "prefect.resource.name" = "{{ body.friendly_name }}"
      "producing-team"        = "Data Science"
    }
  })
}

# Optionally, use a JSON file to load a more complex template
resource "prefect_webhook" "example_with_file" {
  name        = "my-webhook"
  description = "This is a webhook"
  enabled     = true
  template    = file("./webhook-template.json")
}

# Pro / Enterprise customers can assign a Service Account to a webhook to enhance security.
# If set, the webhook request will be authorized with the Service Account's API key.
# NOTE: if the Service Account is deleted, the associated Webhook is also deleted.
resource "prefect_service_account" "authorized" {
  name = "my-service-account"
}
resource "prefect_webhook" "example_with_service_account" {
  name               = "my-webhook-with-auth"
  description        = "This is an authorized webhook"
  enabled            = true
  template           = file("./webhook-template.json")
  service_account_id = prefect_service_account.authorized.id
}

# When importing an existing Webhook resource
# and using a `file()` input for `template`,
# you may encounter inconsequential plan update diffs
# due to minor whitespace changes. This is because
# Terraform's `file()` input does not perform any encoding
# to normalize the input. If this is a problem for you, you
# can wrap the file input in a jsonencode/jsondecode call:
resource "prefect_webhook" "example_with_file_encoded" {
  name        = "my-webhook"
  description = "This is a webhook"
  enabled     = true
  template    = jsonencode(jsondecode(file("./webhook-template.json")))
}

# Access the endpoint of the webhook.
output "endpoints" {
  value = {
    example                      = prefect_webhook.example.endpoint
    example_with_file            = prefect_webhook.example_with_file.endpoint
    example_with_service_account = prefect_webhook.example_with_service_account.endpoint
  }
}