# Get the metadata for the desired account role.
data "prefect_account_role" "member" {
  name = "Member"
}

# Manage an account member's role.
# Note: this resource must be imported before it can be managed
# because it cannot be created by Terraform.
resource "prefect_account_member" "test" {
  account_role_id = data.prefect_account_role.member.id
}
