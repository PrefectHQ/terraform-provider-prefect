resource "prefect_workspace" "workspace" {
	handle = "my-workspace"
	name = "my-workspace"
}

resource "prefect_flow" "flow" {
	name = "my-flow"
	workspace_id = prefect_workspace.workspace.id
	tags = ["tf-test"]
}

resource "prefect_deployment" "deployment" {
	name = "%s"
	description = "string"
	workspace_id = prefect_workspace.workspace.id
	flow_id = prefect_flow.flow.id
	entrypoint = "hello_world.py:hello_world"
	tags = ["test"]
}