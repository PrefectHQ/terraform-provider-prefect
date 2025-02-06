# Prefect Work Pools can be imported using the format `name,workspace_id`
terraform import prefect_work_pool.example kubernetes-work-pool,00000000-0000-0000-0000-000000000000

# You can also import by name only if you have a workspace_id set in your provider
terraform import prefect_work_pool.example kubernetes-work-pool
