provider "time" {}
resource "time_rotating" "ninety_days" {
    rotation_days = 90
}

resource "prefect_service_account" "sa_example" {
    name = "my-service-account"
    api_key_expiration = time_rotating.ninety_days.rotation_rfc3339
}
