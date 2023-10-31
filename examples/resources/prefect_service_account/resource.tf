# NON-EXPIRING API KEY
resource "prefect_service_account" "example" {
  name = "my-service-account"
}

# ROTATING API KEY
# Use the hashicorp/time provider to generate a time_rotating resource
provider "time" {}
resource "time_rotating" "ninety_days" {
  rotation_days = 90
}
# Pass the time_rotating resource to the `api_key_expiration` attribute
# in order to automate the rotation of the Service Account key
resource "prefect_service_account" "example" {
  name               = "my-service-account"
  api_key_expiration = time_rotating.ninety_days.rotation_rfc3339
}
