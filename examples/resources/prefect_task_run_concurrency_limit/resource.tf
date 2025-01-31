provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_task_run_concurrency_limit" "test" {
  workspace_id      = data.prefect_workspace.test.id
  concurrency_limit = 1
  tag               = "test-tag"
}

# Example of a task that will be limited to 1 concurrent run:
/*
from prefect import flow, task

# This task will be limited to 1 concurrent run
@task(tags=["test-tag"])
def my_task():
    print("Hello, I'm a task")


@flow
def my_flow():
    my_task()


if __name__ == "__main__":
    my_flow()
*/
