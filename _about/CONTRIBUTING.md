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
make build
```

The binary will be stored at the root of the project - obtain the absolute path of the project's `./build` directory, as you will need it in the next step

```shell
echo $(pwd)/build
/Users/johnsmith/code/terraform-provider-prefect/build
```

To aid local development, we can use [development overrides for Terraform provider configurations](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) - place this into your `~/.terraformrc` file

```terraform
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "hashicorp/prefect" = "/Users/johnsmith/code/terraform-provider-prefect/build"
  }

  direct {}
}
```

With development overrides, `terraform init` will still initialize the dependency lock, but `terraform apply` commands will disregard the lockfile + use the executable located in the path you specify here (which is keyed off by the provider name).

Note that with `dev_overrides`, you do not need a `required_providers` block

If you ever want to start fresh, go ahead and run:

```shell
make clean
```

though in general, running `make install` will be sufficient in the course of development.

## Testing

There are a few options for running tests depending on the type.

### Unit tests

The following command will run any regular unit tests. These are typically for helper or utility logic, such as data flatteners or equality checks.

```shell
make test
```

### Terraform acceptance tests

The following command will run [TF acceptance tests](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests), by prefixing the test run with `TF_ACC=1`.

```shell
make testacc
```

**NOTE:** Acceptance tests create real Prefect Cloud resources, and require a Prefect Cloud account when running locally

**NOTE:** In most development/contribution cases, acceptance tests will be run in CI/CD via Github Actions, as test-specific credentials are stored in the environment there.  However, if there are instances where a developer wishes to run the tests locally - they can initialize their test provider through the normal environment variables, pointed to an account that they own:

```shell
export PREFECT_API_URL=https://api.prefect.cloud
export PREFECT_API_KEY=<secret>
export PREFECT_CLOUD_ACCOUNT_ID=<uuid>

make testacc
```

### Integration tests

You can also test against a local instance of Prefect. An example of this setup using Docker Compose is available in the [Terraform Provider tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).

First, you'll need to create or modify `~/.terraformrc` on your machine:

```terraform
provider_installation {
  dev_overrides {
    "registry.terraform.io/prefecthq/prefect" = "/Users/<username>/go/bin/"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

You only need to do this once, but if you will need to comment this out any time you want to use the provider from the official Terraform registry instead.

Next, start the Prefect server:

```shell
docker-compose up -d
```

You can confirm the server is running by either:

1. Checking the logs with `docker-compose logs -f`, or
2. Navigating to the UI in your browser at [localhost:4200](http://localhost:4200).

When you're ready to test your changes, compile the provider and install it to your path:

```shell
go install .
```

You can now run `terraform plan` and `terraform apply` to test features in the provider.

## Build Documentation

This provider repository uses the [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) CLI utility to generate markdown documentation.

```shell
make docs
```

The `tfplugindocs` CLI will:

1. Parse all `Schema` objects for the provider, datasources, and resources
2. Create and populate `.md` files for each page of documentation for the objects mentioned in (1)
3. Crawl and extract all named examples in `examples/**` + add those HCL configurations into the examples section of each `.md`

**NOTE:** If any documentation input files inside `examples/**` are modified, Github Actions will automatically run `make docs` and push any udpates to the working branch

## Development considerations

Here are some considerations to keep in mind when developing for the provider.

### Prefect Cloud endpoints

The provider can be configured to target a Prefect Cloud instance, or a self-hosted Prefect instance
using the `endpoint` field.

Prefect Cloud API endpoints often require a `workspace_id` to be configured, either on the provider or on the resource itself.
A helper function named `validateCloudEndpoint` in the `internal/client` package can be used in each client creation method
to validate that a `workspace_id` is configured, and provide an informative error if not.

If the API endpoint does not require a `workspace_id`, such as `accounts`, you can omit this helper function.
