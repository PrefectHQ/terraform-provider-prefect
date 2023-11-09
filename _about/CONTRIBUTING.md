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

On MacOS, you can use `brew` and `go install` to install necessary dependencies:

```shell
brew install \
  go \
  gotestsum \
  golangci-lint \
  goreleaser

go install golang.org/x/tools/cmd/goimports@latest
```

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
    "prefecthq/prefect" = "/Users/johnsmith/code/terraform-provider-prefect/build"
  }

  direct {}
}
```

With development overrides, `terraform init` will still initialize the dependency lock, but `terraform apply` commands will disregard the lockfile + use the executable located in the path you specify here (which is keyed off by the provider name).

If you ever want to start fresh, go ahead and run:

```shell
make clean
```

though in general, running `make install` will be sufficient in the course of development.

## Testing

There are two `make` commands regarding automated tests:

```shell
make test
```

will run any regular unit tests. These are typically for helper or utility logic, such as data flatteners or equality checks.

```shell
make testacc
```

will run [TF acceptance tests](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests), by prefixing the test run with `TF_ACC=1`

**NOTE:** Acceptance tests create real Prefect Cloud resources, and require a Prefect Cloud account when running locally

**NOTE:** In most development/contribution cases, acceptance tests will be run in CI/CD via Github Actions, as test-specific credentials are stored in the environment there.  However, if there are instances where a developer wishes to run the tests locally - they can initialize their test provider through the normal environment variables, pointed to an account that they own:

```shell
export PREFECT_API_URL=https://api.prefect.cloud
export PREFECT_CLOUD_API_KEY=<secret>
export PREFECT_CLOUD_ACCOUNTID=<uuid>

make testacc
```

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
