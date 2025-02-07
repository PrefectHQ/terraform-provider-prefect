# Prefect Webhooks can be imported using the webhook_id
terraform import prefect_webhook.example 11111111-1111-1111-1111-111111111111

# or from a different workspace via the webhook_id,workspace_id
terraform import prefect_webhook.example 11111111-1111-1111-1111-111111111111,00000000-0000-0000-0000-000000000000
