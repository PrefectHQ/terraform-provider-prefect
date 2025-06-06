---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "prefect_team_access Resource - prefect"
subcategory: ""
description: |-
  The resource team_access grants access to a team for a user or service account. For more information, see manage teams https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams.
  This feature is available in the following product plan(s) https://www.prefect.io/pricing: Prefect Cloud (Enterprise).
---

# prefect_team_access (Resource)

The resource `team_access` grants access to a team for a user or service account. For more information, see [manage teams](https://docs.prefect.io/v3/manage/cloud/manage-users/manage-teams).

This feature is available in the following [product plan(s)](https://www.prefect.io/pricing): Prefect Cloud (Enterprise).

## Example Usage

```terraform
# Example: granting access to a service account.

resource "prefect_service_account" "test" {
  name = "my-service-account"
}

resource "prefect_team" "test" {
  name        = "my-team"
  description = "test-team-description"
}

resource "prefect_team_access" "test" {
  member_type     = "service_account"
  member_id       = prefect_service_account.test.id
  member_actor_id = prefect_service_account.test.actor_id
  team_id         = prefect_team.test.id
}


# Example: granting access to a user.

data "prefect_account_member" "test" {
  email = "marvin@prefect.io"
}

resource "prefect_team" "test" {
  name        = "my-team"
  description = "test-team-description"
}

resource "prefect_team_access" "test" {
  team_id         = prefect_team.test.id
  member_type     = "user"
  member_id       = data.prefect_account_member.test.user_id
  member_actor_id = data.prefect_account_member.test.actor_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `member_actor_id` (String) Member Actor ID (UUID)
- `member_id` (String) Member ID (UUID)
- `member_type` (String) Member Type (user | service_account)
- `team_id` (String) Team ID (UUID)

### Optional

- `account_id` (String) Account ID (UUID)

### Read-Only

- `id` (String) Team Access ID
