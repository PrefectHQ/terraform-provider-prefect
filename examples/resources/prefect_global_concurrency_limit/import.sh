# Prefect global concurrency limits can be imported via global_concurrency_limit_id
terraform import prefect_global_concurrency_limit.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via global_concurrency_limit_id,workspace_id
terraform import prefect_global_concurrency_limit.example 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
