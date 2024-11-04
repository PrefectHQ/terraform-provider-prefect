provider "prefect" {}

# All Deployments are scoped to a Workspace.
data "prefect_workspace" "test" {
  handle = "my-workspace"
}

# Be sure to grant all Actors/Teams who need Deployment access to first be
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


# Define the Flow and Deployment, and grant access to the Deployment

resource "prefect_flow" "test" {
  name         = "my-flow"
  workspace_id = data.prefect_workspace.test.id
  tags         = ["test"]
}

resource "prefect_deployment" "test" {
  name         = "my-deployment"
  workspace_id = data.prefect_workspace.test.id
  flow_id      = prefect_flow.test.id
}

resource "prefect_deployment_access" "test" {
  workspace_id  = data.prefect_workspace.test.id
  deployment_id = prefect_deployment.test.id

  manage_actor_ids = [prefect_service_account.test.actor_id]
  run_actor_ids    = [prefect_service_account.test.actor_id]
  view_actor_ids   = [prefect_service_account.test.actor_id]

  manage_team_ids = [data.prefect_team.test.id]
  run_team_ids    = [data.prefect_team.test.id]
  view_team_ids   = [data.prefect_team.test.id]
}