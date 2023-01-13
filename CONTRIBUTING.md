# Contributing 🌳

## Prerequisites

- make
- golang 1.17: `brew install go`

## Developing

Run `make` to see all options for development, test, and publish.

After making any changes to the graphql schema/queries regenerate [api/operations.go](api/operations.go):

```
make api/operations.go
```

To generate docs:

```
make docs
```

To run acceptance tests:

```
 export PREFECT__CLOUD__API_KEY=<your test account api key here>
make testacc
```

To build and install locally into _~/go/bin_:

```
go install
```

To use the locally built version of the provider, use dev overrides to tell terraform where to find it (only needed once):

```
make ~/.terraformrc
```

NB: when using dev overrides `terraform init` will fail, but this can be ignored. If you have more than one provider then they all must be overridden. For more info see [#27459](https://github.com/hashicorp/terraform/issues/27459#issuecomment-1381507253).

## Prefect API

Use Prefect's Interactive API tab to explore GraphQL API. If you can't find what you're looking for in the Documentation Explorer, inspect Chrome Developer Tools whilst taking actions in the interface.

## Resources

- [Terraform Plugin Framework](https://www.terraform.io/plugin/framework) the new way to develop terraform providers. Used by the Prefect provider.
- [hashicorp/terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) template repo for building providers using the plugin framework.
- [hashicorp/terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) for more info on how doc generation works and [Provider Documentation](https://www.terraform.io/registry/providers/docs) for general info about provider documentation.
