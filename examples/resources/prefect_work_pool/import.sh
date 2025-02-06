# Prefect Work Pools can be imported using the name
terraform import prefect_work_pool.example kubernetes-work-pool

# or from a different workspace via name,workspace_id
terraform import prefect_work_pool.example kubernetes-work-pool,00000000-0000-0000-0000-000000000000
