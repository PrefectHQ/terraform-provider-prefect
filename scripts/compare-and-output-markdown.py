# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "inflection",
#     "requests",
#     "tabulate",
# ]
# ///
import requests
from pprint import pprint
from tabulate import tabulate
import inflection
import datetime

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
]

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
    """
    Returns a dict of cloud API resources, with their corresponding
    resource and datasource implementation statuses.

    ex:
    {
      "artifacts": {
        "resource": True,
        "datasource": False,
      },
      "automations": {
        "resource": True,
        "datasource": True,
      },
      ...
    }
    """
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

        implemented_status[cloud_resource] = {
            "resource": is_resource_implemented,
            "datasource": is_datasource_implemented,
        }

    return implemented_status

def generate_wiki_markdown(implemented_status: dict[str, dict[str, bool]]):
    """
    Generates a markdown table of the current implementation status for each cloud resource.

    Outputs this to a `wiki_output.md` file, which can later be used to update the repo wiki.
    """
    headers = ["Cloud Resource", "Resource", "Datasource"]
    markdown_table = tabulate(
        [
            [
                cloud_resource,
                "✅" if status["resource"] else "❌",
                "✅" if status["datasource"] else "❌",
            ]
            for cloud_resource, status in implemented_status.items()
        ],
        headers=headers,
        tablefmt="github",
    )

    with open("wiki_output.md", "w") as f:
        f.write(
            f"_Last updated: {datetime.date.today()}_\n\n{markdown_table}\n"
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

    generate_wiki_markdown(implemented_status)


if __name__ == "__main__":
    main()
