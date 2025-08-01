---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "prefect_workspace_access Resource - prefect"
subcategory: ""
description: |-
  The resource workspace_access represents a connection between an accessor (User, Service Account or Team) with a Workspace Role. This resource specifies an actor's access level to a specific Workspace in the Account.
  Use this resource in conjunction with the workspace_role resource or data source to manage access to Workspaces.
  For more information, see manage workspaces https://docs.prefect.io/v3/manage/cloud/workspaces.
  This feature is available in the following product plan(s) https://www.prefect.io/pricing: Prefect Cloud (Pro), Prefect Cloud (Enterprise).
---

# prefect_workspace_access (Resource)


The resource `workspace_access` represents a connection between an accessor (User, Service Account or Team) with a Workspace Role. This resource specifies an actor's access level to a specific Workspace in the Account.

Use this resource in conjunction with the `workspace_role` resource or data source to manage access to Workspaces.

For more information, see [manage workspaces](https://docs.prefect.io/v3/manage/cloud/workspaces).

This feature is available in the following [product plan(s)](https://www.prefect.io/pricing): Prefect Cloud (Pro), Prefect Cloud (Enterprise).


## Example Usage

```terraform
# ASSIGNING WORKSPACE ACCESS TO A USER
# Read down a default Workspace Role (or create your own)
data "prefect_workspace_role" "developer" {
  name = "Developer"
}

# Read down an existing Account Member by email
data "prefect_account_member" "marvin" {
  email = "marvin@prefect.io"
}

# Assign the Workspace Role to the Account Member
resource "prefect_workspace_access" "marvin_developer" {
  accessor_type     = "USER"
  accessor_id       = prefect_account_member.marvin.user_id
  workspace_id      = "00000000-0000-0000-0000-000000000000"
  workspace_role_id = data.prefect_workspace_role.developer.id
}

# ASSIGNING WORKSPACE ACCESS TO A SERVICE ACCOUNT
# Create a Service Account resource
resource "prefect_service_account" "bot" {
  name = "a-cool-bot"
}

# Assign the Workspace Role to the Service Account
resource "prefect_workspace_access" "bot_developer" {
  accessor_type     = "SERVICE_ACCOUNT"
  accessor_id       = prefect_service_account.bot.id
  workspace_id      = "00000000-0000-0000-0000-000000000000"
  workspace_role_id = data.prefect_workspace_role.developer.id
}

# ASSIGNING WORKSPACE ACCESS TO A TEAM

# Assign the Workspace Role to the Team
resource "prefect_workspace_access" "team_developer" {
  accessor_type     = "TEAM"
  accessor_id       = "11111111-1111-1111-1111-111111111111"
  workspace_id      = "00000000-0000-0000-0000-000000000000"
  workspace_role_id = data.prefect_workspace_role.developer.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `accessor_id` (String) ID (UUID) of accessor to the workspace. This can be an `account_member.user_id` or `service_account.id`
- `accessor_type` (String) USER | SERVICE_ACCOUNT | TEAM
- `workspace_role_id` (String) Workspace Role ID (UUID) to grant to accessor

### Optional

- `account_id` (String) Account ID (UUID) where the workspace is located
- `workspace_id` (String) Workspace ID (UUID) to grant access to

### Read-Only

- `id` (String) Workspace Access ID (UUID)

## Import

Importing workspace access resources is not supported. Instead, define the
resources as usual. If not present, they will be created.
