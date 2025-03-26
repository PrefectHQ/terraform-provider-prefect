resource "prefect_work_pool" "example" {
  name         = "my-work-pool"
  type         = "kubernetes"
  paused       = false
  workspace_id = "my-workspace-id"
}

# Use a JSON file to load a custom base job template,
# since it will likely be a large JSON object.
resource "prefect_work_pool" "example_with_file" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = file("./base-job-template.json")
}

# When importing an existing Work Pool resource
# and using a `file()` input for `base_job_template`,
# you may encounter inconsequential plan update diffs
# due to minor whitespace changes. This is because
# Terraform's `file()` input does not perform any encoding
# to normalize the input. If this is a problem for you, you
# can wrap the file input in a jsonencode/jsondecode call:
resource "prefect_work_pool" "example_with_file_encoded" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = jsonencode(jsondecode(file("./base-job-template.json")))
}

# Alternatively, use the prefect_worker_metadata datasource
# to load a default base job configuration dynamically.
# See the `prefect_worker_metadata` datasource documentation an exhaustive list of keys.
# https://registry.terraform.io/providers/PrefectHQ/prefect/latest/docs/data-sources/worker_metadata
data "prefect_worker_metadata" "d" {}

# example: use the `.kubernetes` key to map to a `kubernetes` work pool type
resource "prefect_work_pool" "example_with_datasource_kubernetes" {
  name              = "test-k8s-pool"
  type              = "kubernetes"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.kubernetes
}

# example: use the `.prefect_managed` key to map to a `prefect:managed` work pool type
resource "prefect_work_pool" "example_with_datasource_prefect_managed" {
  name              = "test-prefect-managed-pool"
  type              = "prefect:managed"
  workspace_id      = data.prefect_workspace.prd.id
  paused            = false
  base_job_template = data.prefect_worker_metadata.d.base_job_configs.prefect_managed
}
