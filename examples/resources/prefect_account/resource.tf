resource "prefect_account" "example" {
  name        = "My Imported Account"
  description = "A cool account"
  settings = {
    allow_public_workspaces = true
    ai_log_summaries        = false
    managed_execution       = false
  }
}
