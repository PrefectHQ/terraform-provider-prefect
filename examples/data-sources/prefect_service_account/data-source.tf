data "prefect_service_account" "bot" {
  id = "00000000-0000-0000-0000-000000000000"
}

# or

data "prefect_service_account" "bot" {
  name = "my-bot-name"
}
