# Prefect Workspaces can be imported via name in the form `name/name-of-workspace`
terraform import prefect_workspace.example name/name-of-workspace

# Prefect Workspaces can also be imported via UUID
terraform import prefect_workspace.example workspace-uuid
