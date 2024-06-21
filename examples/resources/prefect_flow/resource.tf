resource "prefect_workspace" "workspace" {
	handle = "my-workspace"
	name = "my-workspace"
}

resource "prefect_flow" "flow" {
	name = "my-flow"
	workspace_id = prefect_workspace.workspace.id
	tags = ["tf-test"]
}