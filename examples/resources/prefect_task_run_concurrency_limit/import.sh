# Prefect task run concurrency limits can be imported via task_run_concurrency_limit_id
terraform import prefect_task_run_concurrency_limit.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via task_run_concurrency_limit_id,workspace_id
terraform import prefect_task_run_concurrency_limit.example 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
