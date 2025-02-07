# Prefect Deployments can be imported via deployment_id
terraform import prefect_deployment.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via deployment_id,workspace_id
terraform import prefect_deployment.example 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
