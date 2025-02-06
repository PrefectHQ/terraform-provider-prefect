# Prefect Service Accounts can be imported by name in the form `name/my-bot-name`
terraform import prefect_service_account.example name/my-bot-name

# or via UUID
terraform import prefect_service_account.example 00000000-0000-0000-0000-000000000000
