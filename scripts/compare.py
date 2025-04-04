import requests
from pprint import pprint
from rich.console import Console
from rich.table import Table
import inflection


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

    response = {"resources": [], "data-sources": []}
    for item in list_of_implemented_items:
        attributes = item["attributes"]
        category = attributes.get("category")

        if category in response and "slug" in attributes:
            response[category].append(normalize_name(attributes["slug"]))

    return response


def normalize_name(text: str) -> str:
    """
    Normalize a resource name to a singular, lowercase string.

    ex:
    "Block Capabilities" -> "block_capability"
    """
    return inflection.singularize(text.replace(" ", "_").lower())


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


def find_diffs(
    implemented_provider_resources: dict[str, list[str]],
    openapi_resource_tags: list[str],
):
    table = Table(
        show_header=True, show_lines=True, title="Cloud Resource - Provider Status"
    )
    table.add_column("cloud resource")
    table.add_column("resource")
    table.add_column("datasource")

    for cloud_resource in openapi_resource_tags:
        resource = implemented_provider_resources["resources"]
        datasource = implemented_provider_resources["data-sources"]

        resource_status = "✅" if cloud_resource in resource else "❌"
        datasource_status = "✅" if cloud_resource in datasource else "❌"

        table.add_row(cloud_resource, resource_status, datasource_status)

    console = Console()
    console.print(table)


def main():
    implemented_provider_resources = get_provider_implemented_resource_slugs()
    pprint(implemented_provider_resources)

    print()
    print()
    print()
    print()

    openapi_resource_tags = get_prefect_cloud_resources()
    pprint(openapi_resource_tags)

    find_diffs(implemented_provider_resources, openapi_resource_tags)


if __name__ == "__main__":
    main()
