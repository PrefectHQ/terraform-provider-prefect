# prefect_variable resources can be imported by the `name/name_of_variable` identifier
terraform import prefect_variable.example name/name_of_variable

# Alternatively, they can be imported by the variable's ID
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000

# Pass an optional, comma-separated value following the identifier
# if you need to import a resource in a different workspace
# from the one that your provider is configured with
# NOTE: you must specify the workspace_id attribute in the addressed resource
#
# name/<variable_name>,<workspace_id>
terraform import prefect_variable.example name/name_of_variable,11111111-1111-1111-1111-111111111111
# <variable_id>,<workspace_id>
terraform import prefect_variable.example 00000000-0000-0000-0000-000000000000,11111111-1111-1111-1111-111111111111

