---
page_title: "Getting started with the provider"
description: |-
  This guide walks through the necessary provider configurations,
  how to generate a Prefect API Key, as well as a brief introduction
  to Prefect RBAC and permissions
---

# Getting Started with the Prefect Cloud provider

## Provider setup

By default, the provider points to Prefect Cloud. You simply need to pass in an API Key and your Prefect Cloud Account ID.

```terraform
# provider.tf / versions.tf

terraform {
  required_providers {
    prefect = {
      source = "prefecthq/prefect"
    }
  }
}

provider "prefect" {
  account_id = "your Account/Organization ID"
  api_key    = "your API Key"
}
```

You can optionally configure your provider through environment variables in the following way, which implicitly injects the same values as the example above. See our [provider documentation](https://registry.terraform.io/providers/PrefectHQ/prefect/latest/docs) for the full attribute schema + related environment variable names.

```shell
export PREFECT_API_KEY="your API Key"
export PREFECT_CLOUD_ACCOUNT_ID="your Account/Organization ID"
```

```terraform
provider "prefect" {}
```

The optional `account_id` and `workspace_id` attributes set default values, so that any subsequent resources will inherit those values. Set a `workspace_id` if your use case calls for managing only a single Workspace

```terraform
provider "prefect" {
  account_id = "your Account/Organization ID"
  workspace_id = "your Workspace ID"
  api_key    = "your API Key"
}
```

## Finding your Account ID

Navigate to your Account settings in Prefect Cloud

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/account-settings-location.png" alt="Account settings location" align="center" width="400">

<br>

From here, you can locate the `Account ID` value that you will pass into the provider's `account_id` attribute

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/account-id-location.png" alt="Account ID location" align="center" width="400">

## Generating an API Key

The provider can be configured with a [Prefect API Key](https://docs.prefect.io/latest/cloud/users/api-keys/), representing either a User or a Service Account

### Service Account Keys

Most production use-cases will call for using a dedicated Service Account's API Key when invoking Terraform, as the API Key can be managed as a team - independent of any one User.

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/service-account-api-key-example.png" alt="Service Account API Key Example" align="center" width="400">

<br>

After creating the Service Account, the API Key will be generated and displayed for you in the UI.
m
### User Keys

API Keys can also be generated to represent a User - look to this option if you want to use the provider to manage a Personal Account, where the Service Account feature is not available.

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/user-api-key-location.png" alt="User API Key Location" align="center" width="400">

<br>

Note that any User API Keys will inherit the Account/Organization Role of the User (eg. `Admin` or `Member`)

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/user-api-key-example.png" alt="User API Key Example" align="center" width="400">


## RBAC + Permissions

For any API Key used in the provider, note that the associated actor (User or Service Account) needs the necessary permissions to manage certain resources, depending on the associated Role.  See our [documentation on RBAC for more detail](https://docs.prefect.io/latest/cloud/users/roles/).

Prefect Cloud offers built-in Roles on the Account (eg. Organization) and Workspace levels, which you can configure in the [web application](https://app.prefect.cloud/).

For example, if you anticipate needing to manage Account-level resources, such as creating Workspaces, Service Accounts, and assigning Account Roles, the API Key's actor will need the `Admin` [Account Role assigned](https://docs.prefect.io/latest/cloud/users/roles/#organization-level-roles).

Similarly, any operations involving management inside a Workspace will require the appropriate [Workspace Role assigned](https://docs.prefect.io/latest/cloud/users/roles/#workspace-level-roles) to the actor.

### Share existing Workspaces with Terraform provider actor

If you are introducing the Terraform provider into an ***existing Prefect account***, it's likely that you will be working with (and possibly importing) existing Workspaces.

Prefect Workspaces are discrete environments with their own access values, so each actor/accessor will need to be invited to a Workspace to participate - the Terraform provider actor is no different.

-> This only applies to existing Workspaces that will be managed by the provider. New Workspaces created by the provider will have already assigned an `Owner` Workspace Role to the provider actor.

This is more likely to impact Service Account actors, as Users will already know whether or not they have access to a particular Workspace.

To invite the Terraform provider actor to an existing Workspace, so that it can be managed in Terraform:

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/workspace-sharing-location.png" alt="Workspace Sharing Location" align="center" width="400">

<br>

Grant your Terraform provider actor (which is a Service Account in this case) the appropriate [Workspace Role](https://docs.prefect.io/latest/cloud/users/roles/#workspace-level-roles), based on your anticipated use case.

<img src="https://raw.githubusercontent.com/PrefectHQ/terraform-provider-prefect/main/docs/images/service-account-share-example.png" alt="Service Account Share Example" align="center" width="400">
