# Get deployment by ID using Terraform ID reference.
data "prefect_deployment" "existing_by_id" {
  id = prefect_deployment.my_existing_deployment.id
}

# Get deployment by ID string.
data "prefect_deployment" "existing_by_id_string" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get deployment by name using Terraform name reference.
data "prefect_deployment" "existing_by_id_string" {
  name = prefect_deployment.my_existing_deployment.name
}

# Get deployment by name string.
data "prefect_deployment" "existing_by_id_string" {
  name = "my_existing_deployment"
}
