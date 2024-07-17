# Read down the default Admin Account Role
data "prefect_account_role" "admin" {
  name = "Admin"
}

# Read down the default Member Account Role
data "prefect_account_role" "member" {
  name = "Member"
}

# Read down the default Owner Account Role
data "prefect_account_role" "owner" {
  name = "Owner"
}