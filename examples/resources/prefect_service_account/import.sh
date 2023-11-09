# Prefect Service Accounts can be imported via name in the form `name/my-bot-name`
terraform import prefect_service_account.example name/my-bot-name

# Prefect Service Accounts can also be imported via UUID
terraform import prefect_service_account.example 00000000-0000-0000-0000-000000000000
