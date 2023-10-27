data "prefect_workspace_role" "developer" {
	name = "Developer"
}
data "prefect_workspace" "prd" {
	id = "<workspace uuid>"
}
resource "prefect_service_account" "bot" {
	name = "a-cool-bot"
}
resource "prefect_workspace_access" "bot_access" {
	accessor_type = "SERVICE_ACCOUNT"
	accessor_id = prefect_service_account.bot.id
	workspace_id = data.prefect_workspace.prd.id
	workspace_role_id = data.prefect_workspace_role.developer.id
}
