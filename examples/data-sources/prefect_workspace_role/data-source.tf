# Read down the default Owner Workspace Role
data "prefect_workspace_role" "owner" {
  name = "Owner"
}

# Read down the default Worker Workspace Role
data "prefect_workspace_role" "worker" {
  name = "Worker"
}

# Read down the default Developer Workspace Role
data "prefect_workspace_role" "developer" {
  name = "Developer"
}

# Read down the default Viewer Workspace Role
data "prefect_workspace_role" "viewer" {
  name = "Viewer"
}

# Read down the default Runner Workspace Role
data "prefect_workspace_role" "runner" {
  name = "Runner"
}
