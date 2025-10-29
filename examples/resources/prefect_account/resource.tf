resource "prefect_account" "example" {
  name        = "My Imported Account"
  description = "A cool account"
  settings = {
    allow_public_workspaces        = true
    ai_log_summaries               = false
    enforce_webhook_authentication = true
    managed_execution              = false
  }
}
