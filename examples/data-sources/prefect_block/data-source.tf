# Get block by ID using Terraform ID reference.
data "prefect_block" "existing_by_id" {
  id = prefect_block.my_existing_block.id
}

# Get block by ID string.
data "prefect_block" "existing_by_id_string" {
  id = "00000000-0000-0000-0000-000000000000"
}

# Get block by type name and name.
data "prefect_block" "existing_by_id_string" {
  block_type_name = "secret"
  name            = "my_existing_block"
}
