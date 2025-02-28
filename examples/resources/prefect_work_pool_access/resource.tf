provider "prefect" {}

# All work pools are scoped to a Workspace.
data "prefect_workspace" "test" {
  handle = "my-workspace"
}

# Be sure to grant all Actors/Teams who need Work Pool access to first be
# invited to the Workspace (with a role).
data "prefect_workspace_role" "developer" {
  name = "Developer"
}


# Example: invite a Service Account to the Workspace and grant it Developer access

resource "prefect_service_account" "test" {
  name = "my-service-account"
}

resource "prefect_workspace_access" "test" {
  accessor_type     = "SERVICE_ACCOUNT"
  accessor_id       = prefect_service_account.test.id
  workspace_role_id = data.prefect_workspace_role.developer.id
  workspace_id      = data.prefect_workspace.test.id
}


# Example: invite a Team to the Workspace and grant it Developer access

data "prefect_team" "test" {
  name = "my-team"
}

resource "prefect_workspace_access" "test_team" {
  accessor_type     = "TEAM"
  accessor_id       = data.prefect_team.test.id
  workspace_role_id = data.prefect_workspace_role.developer.id
  workspace_id      = data.prefect_workspace.test.id
}


# Define the Work Pool and grant access to the Work Pool

resource "prefect_work_pool" "test" {
  name         = "my-work-pool"
  workspace_id = data.prefect_workspace.test.id
}

resource "prefect_work_pool_access" "test" {
  workspace_id   = data.prefect_workspace.test.id
  work_pool_name = prefect_work_pool.test.name

  manage_actor_ids = [prefect_service_account.test.actor_id]
  run_actor_ids    = [prefect_service_account.test.actor_id]
  view_actor_ids   = [prefect_service_account.test.actor_id]

  manage_team_ids = [data.prefect_team.test.id]
  run_team_ids    = [data.prefect_team.test.id]
  view_team_ids   = [data.prefect_team.test.id]
}