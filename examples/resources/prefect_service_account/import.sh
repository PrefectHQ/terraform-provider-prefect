# Prefect Service Accounts can be imported via name in the form `name/my-bot-name`
terraform import prefect_service_account.example name/my-bot-name

# Prefect Service Accounts can also be imported via UUID
terraform import prefect_service_account.example 7de0291d-59d0-4d57-a629-fe47960aa61b
