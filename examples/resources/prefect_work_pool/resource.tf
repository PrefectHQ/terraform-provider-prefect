resource "prefect_work_pool" "example" {
  name   = "My Work Pool"
  type   = "Kubernetes"
  paused = false
}
