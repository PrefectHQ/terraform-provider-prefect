# Explicit account read
data "prefect_account" "my_organization" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Implicit account read, using the account ID from the provider
data "prefect-account" "my_organization" {}
