---
page_title: "Troubleshooting the provider"
description: |-
  This guide provides troubleshooting resources to help identify
  and fix problems when using the Prefect provider.
---

# Troubleshooting the provider

This guide provides troubleshooting resources to help identify
and fix problems when using the Prefect provider.

If you run into a problem that is not identified below, please open
an issue in our [issue tracker](https://github.com/PrefectHQ/terraform-provider-prefect/issues).

## Error 405: Method Not Allowed

The provider may return an error that looks similar to the following:

```
│ Error: Error during create Work Queue
│ 
│   with prefect_work_queue.doc_processing_subflows,
│   on prefect_work_pool.tf line 557, in resource "prefect_work_queue" "doc_processing_subflows":
│  557: resource "prefect_work_queue" "doc_processing_subflows" {
│ 
│ Could not create Work Queue, unexpected error: failed to create work queue: status code=405 Method
│ Not Allowed, error={"detail":"Method Not Allowed"}
```

This is often an indication that Prefect is being [self-hosted](https://docs.prefect.io/v3/manage/self-host)
and is running behind a proxy. It can usually be fixed by setting `FORWARD_ALLOW_IPS` in your networking
provider.

See related issues for more information:
- [#328: Prefect Flows and Deployments 405 Response](https://github.com/PrefectHQ/terraform-provider-prefect/issues/328)
- [#400: 405 on prefect_work_queue creation](https://github.com/PrefectHQ/terraform-provider-prefect/issues/400)
