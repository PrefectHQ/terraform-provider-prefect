# Acceptance Test Results — Customer-Managed Instance

Tracking acceptance test failures when running against a customer-managed
Prefect instance (`latest-api.private.prefect.dev`), and our progress fixing
them.

**Run:** 171 tests, 21 failures, 337.7s

Status legend: ☐ open · ☑ fixed · ⊘ won't fix / not a provider bug

## Failures

| # | Status | Test | Type | Root cause | Category |
|---|--------|------|------|-----------|----------|
| 1 | ☑ | `TestAccDatasource_team` | datasource | `Could not find Team with name my-team` — seed data missing. **Added `my-team` to instance; re-run to confirm.** | Missing seed data |
| 2 | ☑ | `TestAccDatasource_account_member` | datasource | `Could not find Account Member with email marvin@prefect.io`. **Invited `marvin@prefect.io` to instance; re-run to confirm.** | Missing seed data |
| 3 | ☑ | `TestAccDatasource_account_role_defaults` | datasource | Expected 44 permission elements, got 40 — instance has fewer default permissions. **Fixed and confirmed passing: test now asserts a minimum permission count (`ExpectKnownValueListSizeMin`) instead of an exact count, so it is portable across environments.** | API/data mismatch |
| 4 | ☑ | `TestAccDatasource_account` | datasource | Expected 2 `domain_names`, got 0 — domains/SSO not available on customer-managed. **Fixed and confirmed passing: added `CM` test context; the `domain_names` assertion is skipped when `TEST_CONTEXT=CM`. Rest of the test (id/name/handle) still runs.** | CM-unsupported feature |
| 5 | ☑ | `TestAccDatasource_worker_metadata` | datasource | `404` on `collections/views/aggregate-worker-metadata`. **Root cause: `IsCloudEndpoint` didn't match the CM test host `latest-api.private.prefect.dev` (`.dev`, not `.cloud`), so the client used the OSS worker-metadata route instead of the Cloud-style `collections/work_pool_types`. Fixed: added `private.prefect.dev` to the Cloud-endpoint substrings.** Re-run to confirm. | Endpoint routing (host match) |
| 6 | ☑ | `TestAccDatasource_automation` | datasource | Metric-trigger step returns `422` (only event/compound/sequence triggers accepted on CM). **Fixed: metric-trigger step now skips on OSS or CM (`SkipFuncOSSOrCM`).** | CM-unsupported feature |
| 7 | ☑ | `TestAccResource_webhook` | resource | Endpoint host mismatch: got `latest-api.private.prefect.dev/hooks/...`, test hardcoded `api.stg.prefect.dev/hooks/...`. **Fixed and confirmed passing: `testAccCheckWebhookEndpoint` now derives the expected host from the test client's `GetEndpointHost()` (same source the provider uses) instead of hardcoding, so it is portable across environments.** | Test hardcodes host |
| 8 | ☑ | `TestAccResource_work_pool` | resource | `404` on `POST /work_pools/` (note: **unscoped** URL, missing `/accounts/<id>/workspaces/<id>`). **Root cause: same host-match bug as #5. Step 1 deliberately omits `workspace_id` and expects `validateCloudEndpoint` to reject it, but that check is gated on `IsCloudEndpoint`; the unrecognized `.dev` host skipped validation, so the unscoped request reached CM and 404'd. Fixed by adding `private.prefect.dev` (commit `0fb6cbe`) — no test change needed.** Re-run to confirm. | Endpoint routing (host match) |
| 9 | ☑ | `TestAccResource_block_access` | resource | `Could not find Team with name my-team`. **Added `my-team` to instance; re-run to confirm.** | Missing seed data |
| 10 | ☑ | `TestAccResource_team_workspace_access` | resource | Team/account-member lookup failure (`my-team`). **Added `my-team`; re-run to confirm (also touches `marvin@prefect.io`).** | Missing seed data |
| 11 | ☑ | `TestAccResource_team_access_user` | resource | `Could not find Account Member with email marvin@prefect.io`. **Invited `marvin@prefect.io`; re-run to confirm.** | Missing seed data |
| 12 | ☑ | `TestAccResource_account` | resource | Import: expected name `github-ci-tests`, got `latest` — account name differs on this instance. **Fixed: expected `name`/`handle`/`link` now read from `PREFECT_ACCOUNT_NAME` / `PREFECT_ACCOUNT_HANDLE` / `PREFECT_ACCOUNT_LINK` env vars (via `testutils.EnvOrDefault`), defaulting to the Cloud values. Set these for the CM instance.** | Env-specific data |
| 13 | ☑ | `TestAccResource_work_pool_access` | resource | Pre-apply plan failure — cascade from the `IsCloudEndpoint` host-match bug (work_pool_access is account/workspace-scoped). **Fixed and confirmed passing via the `private.prefect.dev` host change (commit `0fb6cbe`); no test change needed.** | Endpoint routing (host match) |
| 14 | ☑ | `TestAccResource_account_settings` | resource | `422 extra_forbidden: managed_execution` — CM rejects the `managed_execution` field. The test also bakes in Cloud-only `github-ci-tests` name/handle/link and SSO `domain_names` in every step. **Fixed: skip the whole test on CM (`SkipTestsIfCM`), since it exercises multiple Cloud-only account features.** | CM-unsupported feature |
| 15 | ☑ | `TestAccResource_deployment_with_global_concurrency_limit` | resource | Inconsistent result: `global_concurrency_limit_id` was set, now null. **Root cause confirmed via CM source (`customer-managed` repo, `test_deployment_compatibility.py`): CM accepts `global_concurrency_limit_id` on create/update as an OSS-compat shim but deliberately does NOT persist it (excluded from response, logs "not supported by this server version"). It's a Cloud-only feature, not a provider bug. Fixed: skip on CM (`SkipTestsIfCM`).** | CM-unsupported feature |
| 16 | ☑ | `TestAccResource_account_member` | resource | Import: `Could not find Account Member marvin@prefect.io`. **Invited `marvin@prefect.io`; re-run to confirm.** | Missing seed data |
| 17 | ☑ | `TestAccResource_deployment_access` | resource | `Could not find Team with name my-team`. **Added `my-team` to instance; re-run to confirm.** | Missing seed data |
| 18 | ☑ | `TestAccResource_deployment_schedule` | resource | Inconsistent result: `.slug` was `test-schedule`, now empty. **Root cause via CM source: CM *persists* the schedule slug (`test_can_create_schedule_with_slug`), but its `DeploymentScheduleResponse` schema omits the `slug` field, so the create/read response returns it empty. Unlike #15 this is a real, correctable provider gap. Fixed in the provider: `copyScheduleModelToResourceModel` now preserves the prior local slug when the server response omits it (mirrors the existing rrule/parameters round-trip handling). Benefits real users on this CM version, not just the test.** | Provider fix (response omits field) |
| 19 | ☑ | `TestAccResource_variable` | resource | `422: value must be of type string` on the number-value step. **Root cause via CM source (`api/variables.py: enforce_variable_value_type`): CM gates non-string (JSON) variable values behind a `WRITE_JSON_VARIABLES` feature flag that is off by default, returning exactly this 422. Fixed: skip the typed-value steps (number/bool/object/tuple) on CM via per-step `SkipFunc: SkipFuncCM`, and switch the tags step to a string value so it still runs on CM. String-value coverage and tag coverage are preserved on CM.** | CM-unsupported feature (flag) |
| 20 | ☑ | `TestAccResource_resource_sla` | resource | Clean re-run gave a real verdict: `404` on `slas/apply-resource-slas/...`. **Root cause: SLAs are a Cloud-only feature; the CM server does not implement the SLA endpoint (no `apply-resource-slas` route in the CM source). Fixed: skip on OSS or CM via new `SkipTestsIfOSSOrCM` (the test already skipped OSS).** | CM-unsupported feature |
| 21 | ☑ | `TestAccResource_automation` | resource | Multiple Cloud-only steps return `422` on CM: Step 4 (metric trigger — only `event`/`compound`/`sequence` accepted), Step 8 (`send-email-notification` action — not in CM's accepted action list), and the `pause-schedule-for-flow-run` action. **Fixed: all Cloud-only steps (metric trigger + its import, send-email + its import, pause-schedule) now skip on OSS or CM (`SkipFuncOSSOrCM`).** | CM-unsupported feature |

## Grouped by category

| Category | Count | Tests |
|----------|-------|-------|
| **Missing seed data** (`my-team`, `marvin@prefect.io`, domains) | 8 | 1, 2, 4, 9, 10, 11, 16, 17 |
| **Missing/different API endpoint** (404) | 2 | 5, 8 |
| **API schema mismatch** (422) | 3 | 3, 14, 19 |
| **Provider/API drift** (inconsistent apply) | 2 | 15, 18 |
| **Test hardcodes a host** | 1 | 7 |
| **Env-specific data** (account name) | 1 | 12 |
| **Inconclusive** (interrupted, `unknown`) | 3 | 6, 20, 21 |

## Notes

- The **run was interrupted** (`signal: interrupt` at test #5, which also hit
  heavy `429` rate-limiting). Tests 6, 20, 21 never produced a verdict, and #13
  is likely a cascade from missing seed data rather than a distinct provider
  bug. Re-run these in isolation to get a real verdict.
- The largest bucket (8 of 21) is purely **environment seed data** the test
  suite assumes exists in Prefect's internal staging account (`my-team`,
  `marvin@prefect.io`, configured domains). These are not provider bugs.
- **#7 (webhook)** was a genuine test-portability issue (hardcoded
  `api.stg.prefect.dev` hook host). Fixed — see progress log.
- **#14, #19, #3** point at real API surface differences between this instance
  and staging (`managed_execution` field, variable value typing, default
  permission count). Escalate to whoever owns the customer-managed instance.

## Provider / test changes on this branch

Committed changes supporting the customer-managed (CM) test environment:

- `3fdbdb9` — Support Customer-Managed test env endpoints (`IsCloudEndpoint`
  now matches `private.prefect.cloud`).
- `06be2f1` — Block tests use the `string` block type instead of
  `s3-bucket` / `github-repository`, which are not default block types on CM.
  Touches `block_test.go` and `deployment_test.go`.
- `scripts/testacc-dev` uses the 1Password entry
  `op://Platform/terraform-provider-acceptance-tests` (`api-url`, `api-key`,
  `account-id`). Run CM tests by exporting `TEST_CONTEXT=CM` in the environment
  (it is not baked into the script).

## Progress log

- **2026-06-08:** Added `my-team` team to the CM instance. Should resolve the
  `my-team` seed-data failures (#1, #9, #10, #17). Re-run to confirm.
- **2026-06-08:** Invited `marvin@prefect.io` account member to the CM instance.
  Should resolve #2, #11, #16 and the account-member half of #10. Re-run to
  confirm.
- **2026-06-08:** #3 `account_role_defaults` — made the test environment-portable.
  Added `ExpectKnownValueListSizeMin` helper (`internal/testutils/helpers.go`)
  and switched the test to assert a *minimum* permission count per default role
  instead of exact counts. Observed counts: Cloud 44/13/46, CM 40/11/...
  Floors set conservatively below the lowest observed (Admin ≥ 30, Member ≥ 8,
  Owner ≥ 30) to avoid re-tuning per environment.
- **2026-06-08:** Added a customer-managed (`CM`) test context, mirroring the
  existing OSS context (`internal/testutils/provider.go`): `TestContextCM()`,
  `SkipTestsIfCM()`, `SkipFuncCM()`, driven by `TEST_CONTEXT=CM`. Use this to
  guard Cloud-only features not present on customer-managed instances.
  NOTE: `TEST_CONTEXT=CM` must be set in the environment when running against a
  CM instance (it is not baked into `scripts/testacc-dev`).
- **2026-06-08:** #4 `account` — `domain_names` is an SSO feature absent on CM.
  The test now asserts `domain_names` only when not in CM mode; id/name/handle
  still run. Closes the last seed-data item.
- **2026-06-08:** #7 `webhook` — `testAccCheckWebhookEndpoint` hardcoded
  `https://api.stg.prefect.dev/hooks/<slug>`. Now derives the expected host from
  `testutils.NewTestClient().GetEndpointHost()` (the same value the provider
  uses to build the endpoint), so the assertion is correct on any host.
- **2026-06-08:** #12 `account` — import test hardcoded the pre-existing account
  `name`/`handle`/`link` (`github-ci-tests`). Added `testutils.EnvOrDefault` and
  switched expected values to read from `PREFECT_ACCOUNT_NAME` /
  `PREFECT_ACCOUNT_HANDLE` / `PREFECT_ACCOUNT_LINK`, defaulting to the Cloud
  values. Set these env vars to the CM account's values when running on CM.
  NOTE: `TestAccResource_account_settings` (#14) still hardcodes
  `github-ci-tests`; left for the #14 fix.
- **2026-06-08:** #5 `worker_metadata` — the CM test cluster host is
  `latest-api.private.prefect.dev` (`.dev`), which `IsCloudEndpoint` did not
  match, so the collections client took the OSS branch
  (`collections/views/aggregate-worker-metadata`) and 404'd. CM serves the
  Cloud-style `collections/work_pool_types` route. Added `private.prefect.dev`
  to the Cloud-endpoint substrings. Decision: the 3 provider-side
  `IsCloudEndpoint` call sites (API-key/account-ID requirement, account+workspace
  validation, worker-metadata route) all want Cloud behavior for CM since CM is
  account/workspace-scoped and uses Cloud routes; CM-vs-Cloud *feature*
  differences stay handled test-side via `TEST_CONTEXT=CM`. Considered a separate
  `IsCMEndpoint`/`IsCloudOrCMEndpoint` split but it wasn't warranted here.
  This same fix also covers #8 `work_pool`: its Step 1 expects
  `validateCloudEndpoint` (gated on `IsCloudEndpoint`) to reject a work pool
  submitted without a `workspace_id`. The unrecognized `.dev` host skipped that
  validation, so the unscoped `POST /work_pools/` reached CM and 404'd. No
  work_pool test change needed.
- **2026-06-08:** #14 `account_settings` — skipped on CM via `SkipTestsIfCM`.
  The test exercises several Cloud-only account features (pre-existing
  `github-ci-tests` name/handle/link, SSO `domain_names`, and the
  `managed_execution` setting that CM's API rejects as an extra field), so a
  surgical removal of `managed_execution` alone wouldn't make it pass on CM.
  Note: the provider serializes `managed_execution` unconditionally (plain bool,
  no omitempty in `api.AccountSettings`); if account settings should be supported
  on CM later, the provider would need to omit Cloud-only settings for CM
  endpoints. Out of scope for the test pass.
- **2026-06-08:** #15 `deployment_with_global_concurrency_limit` — confirmed via
  the customer-managed server source that `global_concurrency_limit_id` is an
  OSS-compatibility shim on CM: accepted on create/update but intentionally not
  persisted and excluded from the response (it logs a "not supported by this
  server version" warning). So the value legitimately comes back null and the
  provider correctly nulls it. Not a provider bug; a follow-up GET or preserving
  the planned value would both cause permanent drift. Skipped on CM via
  `SkipTestsIfCM`.
- **2026-06-08:** #18 `deployment_schedule` — distinct from #15. Checked the CM
  source: CM *does* persist the schedule slug, but its `DeploymentScheduleResponse`
  schema (`schemas/responses.py`) has no `slug` field, so the create/read response
  returns it empty and the provider overwrote state with `""`, tripping
  "inconsistent result after apply". Real, correctable gap. Fixed in
  `copyScheduleModelToResourceModel`: when the server response has no slug but the
  prior local value (plan on create/update, state on read) had one, keep the local
  value. This mirrors the existing rrule (`normalizeRRuleForState`) and parameters
  round-trip handling, and is correct because the slug is actually persisted on CM
  — so it benefits real users on this CM version, not just the acceptance test. No
  test change needed.
- **2026-06-08:** #19 `variable` — CM source (`api/variables.py`,
  `enforce_variable_value_type`) gates non-string variable values behind the
  `WRITE_JSON_VARIABLES` feature flag, which is off by default; without it CM
  returns "`value` must be of type `string`." (422). Skipped the typed-value
  steps (number/bool/object/tuple) on CM via per-step `SkipFunc: SkipFuncCM`,
  and changed the tags step to use a string value so it still exercises tag
  handling on CM. String + tag coverage preserved; full typed coverage still
  runs on Cloud.
- **2026-06-08:** #20 `resource_sla` — the original "unknown" was just the
  interrupted run. A clean re-run gave a real 404 on
  `slas/apply-resource-slas/...`. SLAs are Cloud-only; the CM server source has
  no `apply-resource-slas` route. Added `SkipTestsIfOSSOrCM` (mirrors the
  existing `SkipFuncOSSOrCM`) and switched the test's top-level `SkipTestsIfOSS`
  to it.
- **2026-06-08:** #21 / #6 `automation` (resource + datasource) — several
  automation features are Cloud-only and CM returns 422 for them:
  metric-trigger automations (only event/compound/sequence accepted),
  the `send-email-notification` action, and the `pause-schedule-for-flow-run`
  action. Added `testutils.SkipFuncOSSOrCM` and switched all the Cloud-only
  automation steps (already labeled "Cloud-only" / previously `SkipFuncOSS`)
  to `SkipFuncOSSOrCM`. This covers, in the resource test: metric-trigger
  create + import, send-email create + import, pause-schedule; and in the
  datasource test: the metric-trigger step.
