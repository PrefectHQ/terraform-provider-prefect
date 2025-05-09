---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "prefect_service_account Data Source - prefect"
subcategory: ""
description: |-
  Get information about an existing Service Account, by name or ID.
  
  Use this data source to obtain service account-level attributes, such as ID.
  
  For more information, see manage service accounts https://docs.prefect.io/v3/manage/cloud/manage-users/service-accounts.
  This feature is available in the following product plan(s) https://www.prefect.io/pricing: Prefect Cloud (Pro), Prefect Cloud (Enterprise).
---

# prefect_service_account (Data Source)

Get information about an existing Service Account, by name or ID.
<br>
Use this data source to obtain service account-level attributes, such as ID.
<br>
For more information, see [manage service accounts](https://docs.prefect.io/v3/manage/cloud/manage-users/service-accounts).


This feature is available in the following [product plan(s)](https://www.prefect.io/pricing): Prefect Cloud (Pro), Prefect Cloud (Enterprise).

## Example Usage

```terraform
data "prefect_service_account" "bot" {
  id = "00000000-0000-0000-0000-000000000000"
}

# or

data "prefect_service_account" "bot" {
  name = "my-bot-name"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) Account ID (UUID), defaults to the account set in the provider
- `id` (String) Service Account ID (UUID)
- `name` (String) Name of the service account

### Read-Only

- `account_role_name` (String) Account Role name of the service account
- `actor_id` (String) Actor ID (UUID), used for granting access to resources like Blocks and Deployments
- `api_key` (String) API Key associated with the service account
- `api_key_created` (String) Date and time that the API Key was created in RFC 3339 format
- `api_key_expiration` (String) Date and time that the API Key expires in RFC 3339 format
- `api_key_id` (String) API Key ID associated with the service account. NOTE: this is always null for reads. If you need the API Key ID, use the `prefect_service_account` resource instead.
- `api_key_name` (String) API Key Name associated with the service account
- `created` (String) Timestamp of when the resource was created (RFC3339)
- `updated` (String) Timestamp of when the resource was updated (RFC3339)
