---
page_title: "Using the provider with Customer-Managed Prefect"
description: |-
  This guide describes how to use the Prefect provider with a
  Customer-Managed Prefect instance, and which Prefect Cloud features
  are not supported in that environment.
---

# Using the provider with Customer-Managed Prefect

The Prefect provider can be used with a
[Customer-Managed](https://docs-customer-managed.prefect.io/) Prefect
instance in addition to Prefect Cloud and self-hosted Prefect (OSS).

A Customer-Managed instance is account- and workspace-scoped and uses the
same API routes as Prefect Cloud, so most resources and data sources work
the same way. Point the provider at your instance's API URL as you would
for Prefect Cloud:

```hcl
provider "prefect" {
  endpoint   = "https://<your-host>/api"
  api_key    = var.prefect_api_key
  account_id = var.prefect_account_id
}
```

The `endpoint`, `api_key`, and `account_id` (and the corresponding
`PREFECT_API_URL`, `PREFECT_API_KEY`, and `PREFECT_CLOUD_ACCOUNT_ID`
environment variables) are used the same way as for Prefect Cloud.

## Unsupported features

Some Prefect Cloud features are not available on a Customer-Managed
instance. Using them generally results in an API error (often `404 Not
Found` or `422 Unprocessable Entity`), or in a value that is accepted by
the API but not persisted.

The following are not supported when the provider is pointed at a
Customer-Managed instance:

- **Account domain names (SSO).** The `domain_names` attribute on
  `prefect_account` reflects SSO configuration that is not available on
  Customer-Managed instances.
- **Account `managed_execution` setting.** The `managed_execution` field
  under `prefect_account.settings` is rejected by the Customer-Managed
  API.
- **Deployment global concurrency limits.** The
  `global_concurrency_limit_id` attribute on `prefect_deployment` is
  accepted by the API for compatibility but is not persisted, so it will
  not take effect. Use `concurrency_limit` instead.
- **Metric-trigger automations.** Automations using a `metric` trigger are
  not supported. Event, compound, and sequence triggers work as expected.
- **Cloud-only automation actions.** The `send-email-notification` and
  `pause-schedule-for-flow-run` actions are not supported. Use the
  `send-notification` action (backed by a notification block) instead of
  `send-email-notification`.
- **Non-string variable values.** `prefect_variable` values must be
  strings. Number, boolean, object, and tuple values are not supported by
  default.
- **Resource SLAs.** The `prefect_resource_sla` resource is not supported,
  as the SLA endpoint is not implemented.

Refer to the
[Customer-Managed documentation](https://docs-customer-managed.prefect.io/)
for the authoritative list of supported features for your instance, as
support may vary by version.

If you run into an unexpected limitation that is not listed here, please
open an issue in our
[issue tracker](https://github.com/PrefectHQ/terraform-provider-prefect/issues).
