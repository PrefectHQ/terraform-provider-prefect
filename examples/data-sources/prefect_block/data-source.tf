data "prefect_block" "existing_by_id" {
  id = prefect_block.my_existing_block.id
}

data "prefect_block" "existing_by_id_string" {
  id = "00000000-0000-0000-0000-000000000000"
}
