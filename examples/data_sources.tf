data "prefect_work_pool" "test" {
    name = "test"
}

output "workpool" {
    value = data.prefect_work_pool.test
}
