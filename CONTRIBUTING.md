# Contributing 🌳

## Prerequisites

- make
- golang 1.17: `brew install go`

## CICD

To release a new version:

- Make sure you are on the main branch
- Set `$version` to the next version
- Tag and push: `git tag $version && gps --tags`

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

To use the locally built version of the provider, tell terraform where to find it (only needed once):

```
make ~/.terraformrc
```

## Prefect API

Use Prefect's Interactive API tab to explore GraphQL API. If you can't find what you're looking for in the Documentation Explorer, inspect Chrome Developer Tools whilst taking actions in the interface.

## Resources

- [Terraform Plugin Framework](https://www.terraform.io/plugin/framework) the new way to develop terraform providers. Used by the Prefect provider.
- [hashicorp/terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) template repo for building providers using the plugin framework.
- [hashicorp/terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) for more info on how doc generation works and [Provider Documentation](https://www.terraform.io/registry/providers/docs) for general info about provider documentation.
