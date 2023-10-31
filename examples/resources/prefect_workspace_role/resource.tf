resource "prefect_workspace_role" "example" {
  name = "Custom Workspace Role"
  scopes = [
    "manage_blocks",
    "see_flows"
  ]
}
