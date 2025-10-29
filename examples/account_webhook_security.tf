# Example: Configure webhook security control on a Prefect account
#
# This example demonstrates how to enable webhook authentication enforcement
# for a Prefect Cloud account. This requires the account to be imported first,
# as accounts cannot be created via the API.
#
# To use this example:
# 1. First import your account:
#    terraform import prefect_account.example <account-id>
# 2. Then apply the configuration to update settings

terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

resource "prefect_account" "example" {
  name   = "my-account"
  handle = "my-account-handle"

  settings {
    # Enable webhook authentication enforcement
    # When enabled, webhooks must be authenticated to trigger events
    enforce_webhook_authentication = true

    # Other optional settings
    allow_public_workspaces = false
    ai_log_summaries        = true
    managed_execution       = true
  }
}
