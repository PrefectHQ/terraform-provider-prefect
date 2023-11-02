# Prefect Workspaces can be imported via handle in the form `handle/workspace-handle`
terraform import prefect_workspace.example handle/workspace-handle

# Prefect Workspaces can also be imported via UUID
terraform import prefect_workspace.example workspace-uuid
