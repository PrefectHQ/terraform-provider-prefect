data "prefect_work_pool" "test" {
    name = "test"
}

resource "prefect_work_pool" "abc" {
    name = "testpool"
    type = "kubernetes"
    paused = false
}

output "workpool" {
    value = data.prefect_work_pool.test
}
