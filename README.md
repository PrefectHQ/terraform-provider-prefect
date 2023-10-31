# terraform-provider-prefect

Terraform provider for [Prefect 2](https://github.com/PrefectHQ/prefect) and [Prefect Cloud](https://app.prefect.cloud).

**Note**: This provider is currently under active development and may have frequent breaking changes.

## Resources

* Account (`prefect_account`)
* Variable (`prefect_variable`)
* Work Pool (`prefect_work_pool`)
* Workspace (`prefect_workspace`)

## Data Sources

* Account (`prefect_account`)
* Variable (`prefect_variable`)
* Work Pool (`prefect_work_pool`)
* Work Pools (`prefect_work_pools`)
* Workspace (`prefect_workspace`)

## Deployment:
The "examples" folder makes use of this local provider.
By following the instructions below you can get it deployed to a target Prefect Cloud 2.0 account.

### 1. Create Prefect 2.0 Service Account
The creation of a [service account](https://docs.prefect.io/ui/service-accounts/#create-a-service-account) generates a key that we will use to authenticate the terraform provider to Prefect Cloud.
The advantage of a service account is that it is not tied to a user, but directly to the account.

### 2. Configure environment variables
In this step we export 3 environment variables:

PREFECT_API_URL = the Prefect Cloud API endpoint
PREFECT_API_KEY = the authentication key (generated as part of the previous step)
PREFECT_ACCOUNT_ID = the account id (by clicking on your organisation you'll see this in the URL)

Run this on your terminal (replace placeholders):

```
export PREFECT_API_URL="https://api.prefect.cloud/api"
export PREFECT_API_KEY="<YOUR-API-KEY>"
export PREFECT_ACCOUNT_ID="<YOUR-ACCOUNT-ID>"
```
### 3. Build the provider
This builds the providers's binary and move it to the Terraform plugins directory (usually under `~/.terraform.d/plugins/`)

Before building the provider make sure you've set the correct CPU architecture of your machine.
E.g: for MAC M1, use `darwin_arm64`, for MAC Intel use `darwin_amd64`.
```
export CPU_ARCHITECTURE="darwin_arm64"
```
Now run the `make` command to build the provider:
```
make install
```

### 4. Deploy the example infrastructure
This is the final step, which deploys the infrastructure to Prefect Cloud
```
cd examples
terraform init
terraform apply
```

In case of a success output, go to Prefect Cloud and find the workspace named `terraform-workspace`

### 5. Tear down the example infrastructure
While the capability to _destroy_ work queues and blocks is in place, you won't be able to completely destroy the stack because the API endpoint to destroy a workspace hasn't been made available. For now you'll need to go to the Prefect Cloud UI, click on the workspace `terraform-workspace` > workspace settings > (hamburger button on the top right) > delete.

Then manually remove the local terraform state files

```
rm -rf examples/.terraform
rm -rf examples/.terraform.lock.hcl
rm -rf examples/.terraform
```

## Notes & Improvements:
* This is far from complete.
* There are no code _tests_ in place at present (adding them will very likely lead to changes to make the code more robust).
* The provider's implementation of `blocks` has been done generically. For a more granular infrastructure state control, consider changing the BlockDocument's `data` field to the target block type e.g: `s3`, `kubernetes job`. This would generate more work and duplication though.
* The GO Prefect 2.0 API (`prefect_api`) should be moved out of the provider, into its own project.
* Should authentication be handled with temporary tokens?

## Development

This project uses the following tools to build:

* The [Go programming language](https://go.dev/dl/)
* GNU `make` to run the build and tests
* `gotestsum` to generate test reports
* `golangci-lint` for static code analysis
* `goreleaser` to generate release binaries (optional for development)

On MacOS, you can use `brew` to install necessary dependencies:

```shell
brew install \
  go \
  gotestsum \
  golangci-lint \
  goreleaser
```

After installing these dependencies, run `make build` to compile.

You can configure your local Terraform installation to use this provider, rather than downloading the provider from the Terraform Registry, by adding the following configuration to your `~/.terraformrc` file (be sure to update the path as needed):

```terraform
provider_installation {
  dev_overrides {
    "prefecthq/prefect" = "/Users/jawnsy/projects/work/terraform-provider-prefect/build"
  }

  direct {}
}
```

### Testing

`make` commands are configured to invoke unit and acceptance tests

```shell
make test
```

In order to run the suite of Acceptance tests:

```shell
make testacc
```

**Note:** Acceptance tests create real Prefect Cloud resources, and require a Prefect Cloud account when running locally

See [Development Overrides for Provider Developers](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) for details.

## Build Documentation

This provider repository uses the [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) CLI utility to generate markdown documentation.

```shell
make docs
```

The `tfplugindocs` CLI will:

1. Parse all `Schema` objects for the provider, datasources, and resources
2. Create and populate `.md` files for each page of documentation for the objects mentioned in (1)
3. Crawl and extract all named examples in `examples/**` + add those HCL configurations into the examples section of each `.md`
