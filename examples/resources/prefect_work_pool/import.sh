# Prefect Work Pools can be imported using the format `workspace_id,name`
terraform import prefect_work_pool.example 00000000-0000-0000-0000-000000000000,kubernetes-work-pool

# You can also import by name only if you have a workspace_id set in your provider
terraform import prefect_work_pool.example kubernetes-work-pool
