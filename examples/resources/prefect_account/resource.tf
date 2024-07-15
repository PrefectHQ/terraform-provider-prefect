resource "prefect_account" "example" {
  name          = "My Imported Account"
  description   = "A cool account"
  billing_email = "marvin@prefect.io"
  settings = {
    allow_public_workspaces = true
    ai_log_summaries        = false
    managed_execution       = false
  }
}
