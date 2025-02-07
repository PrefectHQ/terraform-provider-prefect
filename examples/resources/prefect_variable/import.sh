# prefect_variable resources can be imported by the `name/name_of_variable` identifier
terraform import prefect_variable.example name/name_of_variable

# or via variable_id
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via name/variable_name,workspace_id
terraform import prefect_variable.example name/name_of_variable,11111111-1111-1111-1111-111111111111

# or from a different workspace via variable_id,workspace_id
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000,11111111-1111-1111-1111-111111111111

