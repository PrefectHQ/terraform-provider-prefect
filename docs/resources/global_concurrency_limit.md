---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "prefect_global_concurrency_limit Resource - prefect"
subcategory: ""
description: |-
  The resource global_concurrency_limit represents a global concurrency limit. Global concurrency limits allow you to control how many tasks can run simultaneously across all workspaces. For more information, see apply global concurrency and rate limits https://docs.prefect.io/v3/develop/global-concurrency-limits.
  This feature is available in the following product plan(s) https://www.prefect.io/pricing: Prefect OSS, Prefect Cloud (Free), Prefect Cloud (Pro), Prefect Cloud (Enterprise).
---

# prefect_global_concurrency_limit (Resource)

The resource `global_concurrency_limit` represents a global concurrency limit. Global concurrency limits allow you to control how many tasks can run simultaneously across all workspaces. For more information, see [apply global concurrency and rate limits](https://docs.prefect.io/v3/develop/global-concurrency-limits).

This feature is available in the following [product plan(s)](https://www.prefect.io/pricing): Prefect OSS, Prefect Cloud (Free), Prefect Cloud (Pro), Prefect Cloud (Enterprise).

## Example Usage

```terraform
provider "prefect" {}

data "prefect_workspace" "test" {
  handle = "my-workspace"
}

resource "prefect_global_concurrency_limit" "test" {
  workspace_id          = data.prefect_workspace.test.id
  name                  = "test-global-concurrency-limit"
  limit                 = 1
  active                = true
  active_slots          = 0
  slot_decay_per_second = 1.5
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `limit` (Number) The maximum number of tasks that can run simultaneously.
- `name` (String) The name of the global concurrency limit.

### Optional

- `account_id` (String) Account ID (UUID)
- `active` (Boolean) Whether the global concurrency limit is active.
- `active_slots` (Number) The number of active slots.
- `slot_decay_per_second` (Number) Slot Decay Per Second (number or null)
- `workspace_id` (String) Workspace ID (UUID)

### Read-Only

- `created` (String) Timestamp of when the resource was created (RFC3339)
- `id` (String) Global concurrency limit ID (UUID)
- `updated` (String) Timestamp of when the resource was updated (RFC3339)

## Import

Import is supported using the following syntax:

```shell
# Prefect global concurrency limits can be imported via global_concurrency_limit_id
terraform import prefect_global_concurrency_limit.example 00000000-0000-0000-0000-000000000000

# or from a different workspace via global_concurrency_limit_id,workspace_id
terraform import prefect_global_concurrency_limit.example 00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000000
```
