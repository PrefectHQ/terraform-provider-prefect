provider "prefect" {}

# All Blocks are scoped to a Workspace
data "prefect_workspace" "my_workspace" {
  handle = "my-workspace"
}
resource "prefect_block" "my_secret" {
  name      = "my-secret"
  type_slug = "secret"
  data = jsonencode({
    "value" : "foobar"
  })
  workspace_id = data.prefect_workspace.my_workspace.id
}

# Be sure to grant all Actors/Teams who need Block access
# to first be invited to the Workspace (with a role).
data "prefect_workspace_role" "developer" {
  name = "Developer"
}

# Example: invite a Service Account to the Workspace
resource "prefect_service_account" "bot" {
  name = "bot"
}
resource "prefect_workspace_access" "bot_developer" {
  accessor_type     = "SERVICE_ACCOUNT"
  accessor_id       = prefect_service_account.bot.id
  workspace_role_id = data.prefect_workspace_role.developer.id
  workspace_id      = data.prefect_workspace.my_workspace.id
}

# Example: invite a User to the Workspace
data "prefect_account_member" "user" {
  email = "user@prefect.io"
}
resource "prefect_workspace_access" "user_developer" {
  accessor_type     = "USER"
  accessor_id       = data.prefect_account_member.user.user_id
  workspace_role_id = data.prefect_workspace_role.developer.id
  workspace_id      = data.prefect_workspace.my_workspace.id
}

# Example: invite a Team to the Workspace
data "prefect_team" "eng" {
  name = "my-team"
}
resource "prefect_workspace_access" "team_developer" {
  accessor_type     = "TEAM"
  accessor_id       = data.prefect_team.eng.id
  workspace_role_id = data.prefect_workspace_role.developer.id
  workspace_id      = data.prefect_workspace.my_workspace.id
}

# Grant all Actors/Teams the appropriate Manage or View access to the Block
resource "prefect_block_access" "custom_access" {
  block_id         = prefect_block.my_secret.id
  manage_actor_ids = [prefect_service_account.bot.actor_id]
  view_actor_ids   = [data.prefect_account_member.user.actor_id]
  manage_team_ids  = [data.prefect_team.eng.id]
  workspace_id     = data.prefect_workspace.my_workspace.id
}

# Optionally, leave all fields empty to use the default access controls
resource "prefect_block_access" "default_access" {
  block_id         = prefect_block.my_secret.id
  manage_actor_ids = []
  view_actor_ids   = []
  manage_team_ids  = []
  view_team_ids    = []
  workspace_id     = data.prefect_workspace.my_workspace.id
}
