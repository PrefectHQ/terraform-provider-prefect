data "prefect_service_account" "bot" {
  id = "7de0291d-59d0-4d57-a629-fe47960aa61b"
}

# or

data "prefect_service_account" "bot" {
  name = "my-bot-name"
}
