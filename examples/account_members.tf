data "prefect_account_member" "marvin" {
  email = "marvin@prefect.io"
}
data "prefect_workspace" "prd" {
  id = "<workspace uuid>"
}
data "prefect_workspace_role" "developer" {
  name = "Developer"
}
resource "prefect_workspace_access" "marvin_developer" {
  accessor_type     = "USER"
  accessor_id       = prefect_account_member.marvin.user_id
  workspace_id      = data.prefect_workspace.prd.id
  workspace_role_id = data.prefect_workspace_role.developer.id
}
