resource "prefect_workspace" "workspace" {
  handle = "my-workspace"
  name   = "my-workspace"
}

resource "prefect_block" "demo_github_repository" {
  name      = "demo-github-repository"
  type_slug = "github-repository"

  data = jsonencode({
    "repository_url" : "https://github.com/foo/bar",
    "reference" : "main"
  })

  workspace_id = prefect_workspace.workspace.id
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
  path                = "./foo/bar"
  paused              = false
  storage_document_id = prefect_block.test_gh_repository.id
  version             = "v1.1.1"
  work_pool_name      = "mitch-testing-pool"
  work_queue_name     = "default"
}

