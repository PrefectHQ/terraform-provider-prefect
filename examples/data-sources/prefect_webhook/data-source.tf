# Query by ID
data "prefect_webhook" "example_by_id" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Query by name
data "prefect_webhook" "example_by_name" {
  name = "my-webhook"
}
