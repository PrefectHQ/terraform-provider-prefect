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