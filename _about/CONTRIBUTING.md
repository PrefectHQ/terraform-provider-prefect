# Contributing to the Provider

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli) v0.14.x
- [Go](https://go.dev/doc/install) 1.20.x (to build the provider)

This project also uses the following tools to build and test:

- The [Go programming language](https://go.dev/dl/)
- GNU `make` to run the build and tests
- `gotestsum` to generate test reports
- `golangci-lint` for static code analysis
- `goreleaser` to generate release binaries (optional for development)
- `go-imports` to automatically update missing or unused imports
- `tomlv` to [validate TOML files](https://github.com/BurntSushi/toml/tree/master/cmd/tomlv)

On MacOS, you can use `brew`, `mise` and `go install` to install necessary dependencies:

```shell
brew install \
  go \
  gotestsum \
  mise

go install golang.org/x/tools/cmd/goimports@latest

mise install
```

We use [`mise`](https://github.com/jdx/mise) to manage as many dependencies as possible based on the supported plugins in its [registry](https://mise.jdx.dev/registry.html).

## Local Development

### `pre-commit`

```shell
pre-commit install
```

Be sure to run `pre-commit install` before starting any development. `pre-commit` will ensure that code will be standardized, without you needing to worry about it!

### Building the provider

Anytime you want to test a local change, run the build command, which creates the provider binary in the `./build` folder

```shell
mise run build
```

The binary will be stored at the root of the project under `./build/`.

If you ever want to start fresh, go ahead and run:

```shell
mise run clean
```

though in general, running `mise run build` will be sufficient in the course of development.

## Testing

There are a few options for running tests depending on the type.

### Unit tests

The following command will run any regular unit tests. These are typically for helper or utility logic, such as data flatteners or equality checks.

```shell
mise run test
```

### Terraform acceptance tests

The following command will run [Terraform acceptance tests](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests) by prefixing the test run with `TF_ACC=1`.

```shell
mise run testacc
```

Note that this _does not_ require building the provider binary with `mise run build`.

Acceptance tests create real Prefect Cloud resources, and require a Prefect Cloud account.

In most development and contribution cases, acceptance tests will be run in CI/CD via Github Actions, as test-specific credentials are stored in the environment there. These tests happen in Prefect's internal staging environment, and a Prefect team member must approve the CI action for it to run.

The tests can optionally be triggered from a developer's machine by specifying the Prefect API url, API key, and account ID:

```shell
export PREFECT_API_URL=https://api.prefect.cloud
export PREFECT_API_KEY=<secret>
export PREFECT_CLOUD_ACCOUNT_ID=<uuid>

mise run testacc
```

All acceptance tests run in an ephemeral Prefect workspace, except tests for Prefect workspaces and accounts.
This means that the target Prefect environment needs to allow for multiple workspaces. If your account lacks this,
contribute tests to your pull request and a Prefect team member will review and approve them to run in our internal
infrastructure.

Here are some general guidelines for writing **datasource** acceptance tests

- Test that the datasource works with each supported identifier, usually `name` and `id`

Here are some general guidelines for writing **resource** acceptance tests

- Test that the resource can be created
- Test that the resource can be updated (either by modifying the inputs to a test fixture, or defining a second test fixture with different values)
- Test that the resource can be imported (with a test for each supported identifier, often `id` and `name`)
- (Optional, preferred) Test that a resource _cannot_ be created without all required fields
- (Optional, preferred) Test that a resource _cannot_ be created with invalid values for a field

Finally, here are some general guidelines for both datasources and resources:

- Use the `testutils.Expect*` helper functions to verify values in the Terraform state for state config checks
- Use the `testutils.GetResourceWorkspaceImportStateID` helper function to get the import ID for a resource for import tests

Reference existing tests in `internal/provider/{datasources,resources}/*_test.go` for examples.

For more information, see the [Terraform testing patterns documentation](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns).

### Manual testing

You can also test provider functionality by running `terraform apply` with handcrafted manifests.

To ensure that `terraform` commands use the locally-built binary, we use [development overrides for Terraform provider configurations](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers). These overrides are provided automatically in [dev.tfrc](../dev.tfrc),

First, build the binary:

```bash
mise run build
```

Note: you will need to build the binary each time you change the source code.

Set up the test directory by running:

```bash
mise run dev-new <resource name>
```

This will:
- Create a new directory: `./dev/<resource name>`
- Create a file for environment variables: `./dev/.envrc`
- Create a file for Terraform configuration: `./dev/<resource name>.tf`

The command to change to this directory will automatically be added to your clipboard.
Use this, or manually change directories, to enter the testing directory.

When you run `terraform` commands, you'll notice a message that the locally-built binary is in use:

```plaintext
$ terraform plan
╷
│ Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI configuration:
│  - prefecthq/prefect in ../../build
```

You can then edit the `<resource name>.tf` file with your desired configuration for testing purposes.

### Local testing

You can also test against a local instance of Prefect. An example of this setup using Docker Compose is available in the [Terraform Provider tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).

Start the Prefect server:

```shell
docker-compose up -d
```

You can confirm the server is running by either:

1. Checking the logs with `docker-compose logs -f`, or
2. Navigating to the UI in your browser at [localhost:4200](http://localhost:4200).

Set up the test directory by following the [manual testing instructions](#manual-testing).
Open `<resource name>.tf` and modify the `provider` block with the following:

```terraform
provider "prefect" {
 endpoint = "http://localhost:4200"
}
 ```

You can now run `terraform plan` and `terraform apply` to test features in the
provider against a local instance of Prefect.

## Build Documentation

This provider repository uses the [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) CLI utility to generate markdown documentation.

```shell
mise run docs
```

The `tfplugindocs` CLI will:

1. Parse all `Schema` objects for the provider, datasources, and resources
2. Create and populate `.md` files for each page of documentation for the objects mentioned in (1)
3. Crawl and extract all named examples in `examples/**` + add those HCL configurations into the examples section of each `.md`

**NOTE:** If any documentation input files inside `examples/**` are modified, Github Actions will automatically run `mise run docs` and push any udpates to the working branch

Documentation will be rendered into the `docs/` directory. If you need to add additional
information to a document page, create a template first. To do this, create a new file under either:

- `templates/resources/<object>.md.tpl`, or
- `templates/data-sources/<object>.md.tpl`

In this new file, copy the following content:

```markdown
---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "{{.Name}} {{.Type}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{.Description}}

## Example Usage

{{tffile .ExampleFile}}

{{.SchemaMarkdown | trimspace}}

## Import

Import is supported using the following syntax:

{{codefile "shell" .ImportFile}}
```

Additional documentation should be added in this template file to be rendered into the
`docs/` directory by `tfplugindocs`.

Any sections that are not applicable can be removed. For example, data sources and some resources
do not support import functionality, and that section can be removed from the template.

## Development considerations

Here are some considerations to keep in mind when developing for the provider.

### Prefect Cloud endpoints

The provider can be configured to target a Prefect Cloud instance, or a self-hosted Prefect instance
using the `endpoint` field.

Prefect Cloud API endpoints often require a `workspace_id` to be configured, either on the provider or on the resource itself.
A helper function named `validateCloudEndpoint` in the `internal/client` package can be used in each client creation method
to validate that a `workspace_id` is configured, and provide an informative error if not.

If the API endpoint does not require a `workspace_id`, such as `accounts`, you can omit this helper function.

### API route considerations

Certain API routes require a traling slash (`/`). These are most often for `POST` methods used
for creating resources.

If the trailing slash is not provided, it can lead to errors such as 404, 405, 307, etc.

A helper script is available in `../scripts/trailing-slash-routes`. Running this will produce JSON output that
lists which routes end with a slash, along with the method and description to more easily identify which functions
to check under `../internal/client/*.go`.

### Test fixture helpers

Test fixtures are used to create resources in the test environment. They are typically used to create resources
that are used in the test, such as a workspace, account, or deployment.

The `internal/provider/testutils` package contains a helper function named `RenderTemplate` that can be used to
create test fixtures that contain multiple resources. This function is especially useful for creating test fixtures that
contain multiple resources, making it easier to visually understand where each variable is inserted when compared to
using `fmt.Sprintf`.

- If the fixture is fairly short and has fewer than 3-5 variables, `fmt.Sprintf` is usually sufficient.
- If the fixture is longer, or has more than 3-5 variables, `RenderTemplate` is a better choice.

For examples for both approaches, see `internal/provider/{resources,datasources}/*_test.go` files.

### Writing tests for Prefect Cloud and Prefect OSS

The [Prefect Cloud API][Prefect Cloud API] and [Prefect OSS API][Prefect OSS API]
have slight difference that we need to account for in the Terraform provider.

To ensure compatibility with both options, we run Terraform acceptance tests against both.

For features that are known to be Prefect Cloud-only, skip the entire feature. For example:

```go
func TestAccResource_automation(t *testing.T) {
	// Automations are not supported in OSS.
	testutils.SkipTestsIfOSS(t)
}
```

For features that work with both Prefect Cloud and Prefect OSS, you will usually need additional
logic to account for Cloud-only resources like workspaces. For example, see the
[`prefect_block` tests](https://github.com/PrefectHQ/terraform-provider-prefect/blob/main/internal/provider/resources/block_test.go):

- The fixture config struct includes fields for `Workspace` and `WorkspaceIDArg`.
- The fixture config function uses these fields in the string.
- Using the `testutils.TestContextOSS` method, the workspace value is excluded if the context is Prefect OSS.
- That same method can be used elsewhere in the testing logic to account for differences between Cloud and OSS.

[Prefect Cloud API]: https://app.prefect.cloud/api/docs
[Prefect OSS API]: https://docs.prefect.io/3.0/api-ref/rest-api
