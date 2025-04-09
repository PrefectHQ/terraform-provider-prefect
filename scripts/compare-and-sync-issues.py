# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "inflection",
#     "requests",
#     "rich",
#     "pygithub",
# ]
# ///
import requests
from pprint import pprint
from rich.console import Console
from rich.table import Table
import inflection
from github import Github
from github import Auth
import os
import json

# Update this dict to whitelist an API resource
# which may not have a direct TF implementation with the same name,
# but is functionally covered by a TF resources/datasources.
#   key = normalized cloud resource name
#   value = list of known TF aliases (resource or datasource)
known_aliases = {
    "account_membership": ["account_member"],
    "account_role": ["account_member"],
    "block_document": ["block"],
    "block_type": ["block"],
    "bot": ["service_account"],
    "concurrency_limits_v2": ["global_concurrency_limit"],
    "notification": ["resource_sla"],
    "sla": ["resource_sla"],
    "workspace_bot_access": ["workspace_access"],
    "workspace_team_access": ["workspace_access"],
    "workspace_user_access": ["workspace_access"],
}

# Update this list to omit an API resource from this reporting.
#   value = normalized cloud resource name
#   (normalized = singularize + underscore)
apis_we_wont_implement = [
    "account_billing",
    "account_sso",
    "ai",
    "block_capability",
    "concurrency_limit",
    "collection",
    "download",
    "event",
    "flow_run_state",
    "flow_run",
    "invitation",
    "log",
    "me",
    "metric",
    "rate_limit",
    "root",
    "savedsearch",
    "schema",
    "task_run_state",
    "task_run",
    "task_worker",
    "ui",
    "workspace_invitation",
    "workspace_scope",
    "flow",
]


def get_latest_tf_provider_version_id():
    url = "https://registry.terraform.io/v2/providers/PrefectHQ/prefect?include=categories,moved-to,potential-fork-of,provider-versions,top-modules&include=categories%2Cmoved-to%2Cpotential-fork-of%2Cprovider-versions%2Ctop-modules&name=prefect&namespace=PrefectHQ"

    response = requests.get(url)
    json = response.json()
    versions = json["data"]["relationships"]["provider-versions"]["data"]
    return versions[-1]["id"]


def get_provider_implemented_resource_slugs() -> dict[str, list[str]]:
    """returns a dict of implemented resources / datasources

    {
      "resources": [
        "account",
        "user",
        ...
      ],
      "data-sources": [
        "accounts",
        "users",
        ...
      ]
    }

    """
    provider_id = get_latest_tf_provider_version_id()
    url = f"https://registry.terraform.io/v2/provider-versions/{provider_id}?include=provider-docs"
    response = requests.get(url)
    json = response.json()

    list_of_implemented_items = json["included"]

    implemented_tf_items: dict[str, list[str]] = {"resources": [], "data-sources": []}
    for item in list_of_implemented_items:
        attributes = item["attributes"]
        category = attributes.get("category")

        if category in implemented_tf_items and "slug" in attributes:
            implemented_tf_items[category].append(normalize_name(attributes["slug"]))

    implemented_tf_items["resources"] = sorted(implemented_tf_items["resources"])
    implemented_tf_items["data-sources"] = sorted(implemented_tf_items["data-sources"])

    return implemented_tf_items


def normalize_name(text: str) -> str:
    """
    Normalize a resource name to a singular, lowercase string.

    ex:
    "Block Capabilities" -> "block_capability"
    """
    return inflection.singularize(text.replace(" ", "_").lower())
    inflection.pluralize


def get_prefect_cloud_resources() -> list[str]:
    """fetches the Cloud openapi.json, parse it, and
    returns a list of unique, sorted resource tags.

    ex:
    [
      'artifacts',
      'automations',
      'block_capabilities',
      'block_documents',
      'block_schemas',
      'block_types',
      'bots',
      ...
    ]
    """
    url = "https://api.prefect.cloud/api/openapi.json"
    response = requests.get(url)
    json = response.json()

    tags = set()
    for _, path in json["paths"].items():
        for _, method in path.items():
            tags.update(method.get("tags", []))
    return [normalize_name(tag) for tag in sorted(tags)]


def does_exist_in_list(resource: str, list: list[str]) -> bool:
    """
    Check if a cloud resource exists in either the TF resource or datasource list.

    This performs a normalized value check by checking:
        - does the normalized cloud resource exist?
        - does the cloud resource have a known alias?
    """
    if resource in list:
        return True

    for value in known_aliases.get(resource, []):
        if value in list:
            return True

    return False


def find_diffs(
    implemented_provider_resources: dict[str, list[str]],
    openapi_resource_tags: list[str],
) -> dict[str, dict[str, bool]]:
    table = Table(
        show_header=True, show_lines=True, title="Cloud Resource - Provider Status"
    )
    table.add_column("cloud resource")
    table.add_column("resource")
    table.add_column("datasource")

    implemented_status: dict[str, dict[str, bool]] = {}
    for cloud_resource in openapi_resource_tags:
        if cloud_resource in apis_we_wont_implement:
            continue

        is_resource_implemented = does_exist_in_list(
            cloud_resource,
            implemented_provider_resources["resources"],
        )
        is_datasource_implemented = does_exist_in_list(
            cloud_resource,
            implemented_provider_resources["data-sources"],
        )

        table.add_row(
            cloud_resource,
            "✅" if is_resource_implemented else "❌",
            "✅" if is_datasource_implemented else "❌",
        )

        implemented_status[cloud_resource] = {
            "resource": is_resource_implemented,
            "datasource": is_datasource_implemented,
        }

    console = Console()
    console.print(table)

    return implemented_status


def sync_github_issues(implemented_status: dict[str, dict[str, bool]]):
    issue_label_to_search = "parity-audit"

    auth = Auth.Token(os.environ["GITHUB_TOKEN"])
    g = Github(auth=auth)

    repo = g.get_repo("prefecthq/terraform-provider-prefect")
    existing_parity_audit_issues = repo.get_issues(
        state="open", labels=[issue_label_to_search]
    )

    issue_title_prefix: str = "Feature Request: "

    # close existing feature request issues which have later been determined to not be implemented
    # or were deemed to be implemented after an addition to the known_aliases list
    for issue in existing_parity_audit_issues:
        requested_api_resource = issue.title.removeprefix(issue_title_prefix)

        if requested_api_resource in apis_we_wont_implement:
            print(
                f"{requested_api_resource}: closing feature request due to not being implemented"
            )
            issue.edit(state="closed", state_reason="not_planned")
        elif (
            requested_api_resource in implemented_status
            and implemented_status[requested_api_resource]["resource"]
            and implemented_status[requested_api_resource]["datasource"]
        ):
            print(
                f"{requested_api_resource}: closing feature request due to being implemented"
            )
            issue.edit(state="closed", state_reason="completed")

    # upsert feature request issues, based on our dictionary of resources to implement
    for resource_to_check, status in implemented_status.items():
        # skip if both resource and datasource are implemented
        if status["resource"] and status["datasource"]:
            continue

        issue_body = f"Implementation Status: `{json.dumps(status)}`"

        # if the issue already exists, update the body with the current status
        if any(
            resource_to_check == issue.title.removeprefix(issue_title_prefix)
            for issue in existing_parity_audit_issues
        ):
            issue.edit(body=issue_body)

        # if the issue doesn't exist, let's create it
        else:
            repo.create_issue(
                title=f"{issue_title_prefix}{resource_to_check}",
                body=issue_body,
                labels=[issue_label_to_search],
            )


def main():
    implemented_provider_resources = get_provider_implemented_resource_slugs()
    pprint(implemented_provider_resources)

    print()
    print()
    print()
    print()

    openapi_resource_tags = get_prefect_cloud_resources()
    pprint(openapi_resource_tags)

    implemented_status = find_diffs(
        implemented_provider_resources,
        openapi_resource_tags,
    )

    sync_github_issues(implemented_status)


if __name__ == "__main__":
    main()
