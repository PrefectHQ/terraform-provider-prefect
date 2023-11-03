resource "prefect_work_pool" "example" {
  name         = "my-work-pool"
  type         = "kubernetes"
  paused       = false
  workspace_id = "my-workspace-id"
}

# Use a JSON file to load a base job configuration
resource "prefect_work_pool" "example" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = file("./base-job-template.json")
}

# Or use the prefect_worker_metadata datasource
# to load a default base job configuration
data "prefect_worker_metadata" "d" {}

resource "prefect_work_pool" "example" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.kubernetes
}
