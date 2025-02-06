# Prefect Webhooks can be imported using the format `id,workspace_id`
terraform import prefect_webhook.example 11111111-1111-1111-1111-111111111111,00000000-0000-0000-0000-000000000000

# You can also import by id only if you have a workspace_id set in your provider
terraform import prefect_webhook.example 00000000-0000-0000-0000-000000000000
