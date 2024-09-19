resource "prefect_workspace" "workspace" {
  handle = "my-workspace"
  name   = "my-workspace"
}

resource "prefect_flow" "flow" {
  name         = "my-flow"
  workspace_id = prefect_workspace.workspace.id
  tags         = ["tf-test"]
}

resource "prefect_deployment" "deployment" {
  name                     = "my-deployment"
  description              = "string"
  workspace_id             = prefect_workspace.workspace.id
  flow_id                  = prefect_flow.flow.id
  entrypoint               = "hello_world.py:hello_world"
  tags                     = ["test"]
  enforce_parameter_schema = false
  job_variables = jsonencode({
    "env" : { "some-key" : "some-value" }
  })
  manifest_path = "./bar/foo"
  parameters = jsonencode({
    "some-parameter" : "some-value",
    "some-parameter2" : "some-value2"
  })
  parameter_openapi_schema = jsonencode({
    "type" : "object",
    "properties" : {
      "some-parameter" : { "type" : "string" }
    }
  })
  path            = "./foo/bar"
  paused          = false
  version         = "v1.1.1"
  work_pool_name  = "mitch-testing-pool"
  work_queue_name = "default"
}

