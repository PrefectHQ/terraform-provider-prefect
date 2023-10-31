resource "prefect_account" "example" {
  name                    = "My Imported Account"
  description             = "A cool account"
  billing_email           = "marvin@prefect.io"
  allow_public_workspaces = true
}
