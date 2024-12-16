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

# Use a JSON file to load a more complex template
resource "prefect_webhook" "example_with_file" {
  name        = "my-webhook"
  description = "This is a webhook"
  enabled     = true
  template    = file("./webhook-template.json")
}

# Access the endpoint of the webhook.
output "endpoint" {
  value = prefect_webhook.example_with_file.endpoint
}
