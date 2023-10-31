# ASSIGNING WORKSPACE ACCESS TO A USER
# Read down a default Workspace Role (or create your own)
data "prefect_workspace_role" "developer" {
  name = "Developer"
}

# Read down an existing Account Member by email
data "prefect_account_member" "marvin" {
  email = "marvin@prefect.io"
}

# Assign the Workspace Role to the Account Member
resource "prefect_workspace_access" "marvin_developer" {
  accessor_type     = "USER"
  accessor_id       = prefect_account_member.marvin.user_id
  workspace_id      = "workspace-UUID"
  workspace_role_id = data.prefect_workspace_role.developer.id
}

# ASSIGNING WORKSPACE ACCESS TO A SERVICE ACCOUNT
# Create a Service Account resource
resource "prefect_service_account" "bot" {
  name = "a-cool-bot"
}

# Assign the Workspace Role to the Service Account
resource "prefect_workspace_access" "bot_developer" {
  accessor_type     = "SERVICE_ACCOUNT"
  accessor_id       = prefect_service_account.bot.id
  workspace_id      = "workspace-UUID"
  workspace_role_id = data.prefect_workspace_role.developer.id
}
