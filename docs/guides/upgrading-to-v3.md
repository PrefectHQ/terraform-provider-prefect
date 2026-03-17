---
page_title: "Upgrading to v3"
description: |-
  This guide walks through breaking changes introduced in v3 of the
  Prefect Terraform provider, and the steps needed to update your
  configuration.
---

# Upgrading to v3

This guide covers breaking changes in v3 of the Prefect Terraform provider and how to update your configuration.

## Removed deprecated attributes

Several attributes that were deprecated in v2 have been removed in v3. If your configuration references any of these attributes, you will need to remove them before upgrading.

### `prefect_deployment` resource and data source

The `manifest_path` attribute has been removed. This field was unused by Prefect Cloud and was only preserved for backward compatibility with older clients.

**Action required:** Remove any `manifest_path` references from your `prefect_deployment` resource and data source configurations.

```terraform
resource "prefect_deployment" "example" {
  name         = "my-deployment"
  flow_id      = prefect_flow.example.id
  workspace_id = prefect_workspace.example.id

  # Remove this line:
  # manifest_path = "path/to/manifest.json"
}
```

### `prefect_deployment_schedule` resource

The `max_active_runs` and `catchup` attributes have been removed. These fields were Cloud-only and are no longer used. Concurrency is now managed through [concurrency limits](https://docs.prefect.io/v3/deploy/index#concurrency-limiting).

**Action required:** Remove any `max_active_runs` or `catchup` references from your `prefect_deployment_schedule` resource configurations.

```terraform
resource "prefect_deployment_schedule" "example" {
  deployment_id = prefect_deployment.example.id
  workspace_id  = prefect_workspace.example.id
  active        = true
  cron          = "0 0 * * *"

  # Remove these lines:
  # max_active_runs = 10
  # catchup         = true
}
```

### `prefect_account` resource and data source

The `billing_email` attribute has been removed. This field is managed by Stripe and was never writable through the Prefect API.

**Action required:** Remove any `billing_email` references from your `prefect_account` resource and data source configurations, including any outputs or references to this attribute.

```terraform
resource "prefect_account" "example" {
  name   = "my-organization"
  handle = "my-org"

  # Remove this line:
  # billing_email = "billing@example.com"
}
```

## Upgrade steps

1. **Update your configuration.** Search your `.tf` files for any references to `manifest_path`, `max_active_runs`, `catchup`, or `billing_email` and remove them.

2. **Update the provider version constraint.**

   ```terraform
   terraform {
     required_providers {
       prefect = {
         source  = "prefecthq/prefect"
         version = "~> 3.0"
       }
     }
   }
   ```

3. **Run `terraform init -upgrade`** to download the new provider version.

4. **Run `terraform plan`** to verify that your configuration is valid and that no unexpected changes are detected.

If you run into issues not covered here, please open an issue in our [issue tracker](https://github.com/PrefectHQ/terraform-provider-prefect/issues).
