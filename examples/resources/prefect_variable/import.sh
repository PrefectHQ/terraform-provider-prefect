# Prefect Variables can be imported via name in the form `name/name-of-variable`
terraform import prefect_variable.example name/name-of-variable

# Prefect Variables can also be imported via UUID
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000
