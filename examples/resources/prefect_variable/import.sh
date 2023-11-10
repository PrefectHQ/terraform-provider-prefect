# Prefect Variables can be imported via name in the form `name/name_of_variable`
terraform import prefect_variable.example name/name_of_variable

# Prefect Variables can also be imported via UUID
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000
