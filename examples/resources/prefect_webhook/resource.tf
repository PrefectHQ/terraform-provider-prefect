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

# For dynamic webhook templates where the desired payload should be
# in JSON format, use the heredoc format to ensure the expression
# is preserved.
#
# As an example, this webhook could be called via curl:
#
# ```bash
# curl -X POST "${url}" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "foo": "foo",
#     "bar": "bar"
#   }'
#
# For more information, see:
# - https://developer.hashicorp.com/terraform/language/expressions/strings#heredoc-strings
# - https://docs.prefect.io/v3/concepts/webhooks#dynamic-webhook-events
resource "prefect_webhook" "example_with_dynamic_template" {
  name        = "my-webhook"
  description = "This is a webhook"
  enabled     = true

  template = <<-EOF
    {
      "event": "test.body.passthrough",
      "resource": {
          "prefect.resource.id": "test.body-passthrough",
          "prefect.resource.name": "body-passthrough"
      },
      "payload": {{ body | tojson }}
    }
  EOF
}

# Webhook templates don't need to be JSON. You can use a bare Jinja
# expression as the template string, such as the Cloud Events helper:
# - https://docs.prefect.io/v3/automate/events/webhook-triggers#webhook-templates
resource "prefect_webhook" "example_with_string_template" {
  name        = "my-cloud-events-webhook"
  description = "Receives CloudEvents-formatted webhooks"
  enabled     = true
  template    = "{{ body|from_cloud_event(headers) }}"
}

# Access the endpoint of the webhook.
output "endpoints" {
  value = {
    example                      = prefect_webhook.example.endpoint
    example_with_file            = prefect_webhook.example_with_file.endpoint
    example_with_service_account = prefect_webhook.example_with_service_account.endpoint
  }
}
