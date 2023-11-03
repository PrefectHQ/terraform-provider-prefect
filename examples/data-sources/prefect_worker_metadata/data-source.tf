# Use the prefect_worker_metadata datasource
# to fetch a set of default base job configurations
# to be used with several common worker types.
data "prefect_worker_metadata" "d" {}

resource "prefect_work_pool" "kubernetes" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.kubernetes
}

resource "prefect_work_pool" "ecs" {
  name              = "test-ecs-pool"
  type              = "ecs"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.ecs
}

resource "prefect_work_pool" "process" {
  name              = "test-process-pool"
  type              = "cloud-run:push"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.cloud_run_push
}
