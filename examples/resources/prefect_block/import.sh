# prefect_block resources can be imported by the block_id
terraform import prefect_block.my_block 00000000-0000-0000-0000-000000000000
#
# or from a different workspace via block_id,workspace_id
terraform import prefect_block.my_block 00000000-0000-0000-0000-000000000000,11111111-1111-1111-1111-111111111111

