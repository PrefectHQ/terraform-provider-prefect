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
  parameters = jsonencode({
    "some-parameter" : "some-value",
    "some-parameter2" : "some-value2"
  })
  parameter_openapi_schema = jsonencode({
    "type" : "object",
    "properties" : {
      "some-parameter" : { "type" : "string" }
      "some-parameter2" : { "type" : "string" }
    }
  })
  path   = "./foo/bar"
  paused = false
  pull_steps = [
    {
      type      = "set_working_directory",
      directory = "/some/directory",
    },
    {
      type               = "git_clone"
      repository         = "https://github.com/some/repo"
      branch             = "main"
      include_submodules = true

      # For private repositories, choose from one of the following options:
      #
      # Option 1: using an access token by passing it as plaintext
      access_token = "123abc"
      # Option 2: using an access token by referencing a Secret block
      access_token = "{{ prefect.blocks.secret.github-token }}"
      # Option 3: using a Credentials block
      credentials = "{{ prefect.blocks.github-credentials.private-repo-creds }}"
    },
    {
      type      = "pull_from_azure_blob_storage",
      requires  = "prefect-azure[blob_storage]"
      container = "my-container",
      folder    = "my-folder",
    },
    {
      type     = "pull_from_s3",
      requires = "prefect-aws>=0.3.4"
      bucket   = "some-bucket",
      folder   = "some-folder",
    }
  ]
  storage_document_id = prefect_block.test_gh_repository.id
  version             = "v1.1.1"
  work_pool_name      = "some-testing-pool"
  work_queue_name     = "default"
}

