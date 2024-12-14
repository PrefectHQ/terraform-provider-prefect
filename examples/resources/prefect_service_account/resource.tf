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
resource "prefect_service_account" "example_rotate_time_key" {
  name               = "my-service-account"
  api_key_expiration = time_rotating.ninety_days.rotation_rfc3339
}

# Optionally, rotate non-expiring Service Account keys
# using the `api_key_keepers` attribute, which is an
# arbitrary map of values that, if changed, will
# trigger a key rotation (but not a re-creation of the Service Account)
resource "prefect_service_account" "example_rotate_forever_key" {
  name               = "my-service-account"
  api_key_expiration = null # never expires
  api_key_keepers = {
    foo = "value-1"
    bar = "value-2"
  }
}

# Use the optional `old_key_expires_in_seconds`, which applies
# a TTL to the old key when rotation takes place.
# This is useful to ensure that your consumers don't experience
# downtime when the new key is being rolled out.
resource "prefect_service_account" "example_old_key_expires_later" {
  name                       = "my-service-account"
  old_key_expires_in_seconds = 300

  # Remember that `old_key_expires_in_seconds` is only applied
  # when a key rotation takes place, such as changing the
  # `api_key_expiration` attribute
  api_key_expiration = time_rotating.ninety_days.rotation_rfc3339

  # or the `api_key_keepers` attribute
  api_key_keepers = {
    foo = "value-1"
    bar = "value-2"
  }
}
