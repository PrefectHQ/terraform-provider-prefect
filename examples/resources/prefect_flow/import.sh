# Prefect Flows can be imported via flow_id
terraform import prefect_flow.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via flow_id,workspace_id
terraform import prefect_flow.example 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
