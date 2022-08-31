MAKEFLAGS += --warn-undefined-variables
SHELL = /bin/bash -o pipefail
.DEFAULT_GOAL := help
.PHONY: help fmt apply state lint fix tidy test testacc vet

GOBIN := $(shell go env GOPATH)/bin

## display help message
help:
	@awk '/^##.*$$/,/^[~\/\.0-9a-zA-Z_-]+:/' $(MAKEFILE_LIST) | awk '!(NR%2){print $$0p}{p=$$0}' | awk 'BEGIN {FS = ":.*?##"}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

## generate API models from graphql schema and queries
api/operations.go: api/operations.graphql api/common.graphql api/projects.graphql api/service_account.graphql
	go run git.sr.ht/~emersion/gqlclient/cmd/gqlclientgen@latest -q api/operations.graphql -s api/common.graphql -s api/projects.graphql -s api/service_account.graphql -o api/operations.go
# update these scalars to a type alias, because these types are lowercase and so aren't exported and so can't be referenced for casting
	sed -i '' 's/type timestamptz string/type timestamptz = string/' api/operations.go
	sed -i '' 's/type uuid string/type uuid = string/' api/operations.go
	sed -i '' 's/type membership_role string/type membership_role = string/' api/operations.go
# fix NotAType error because the name of the return value is the same as a type and gqclient doesn't support field aliases
	sed -i '' 's/(user_view_same_tenant \[\]user_view_same_tenant, err error)/(tenant_user \[\]user_view_same_tenant, err error)/' api/operations.go
	sed -i '' 's/(auth_role \[\]auth_role, err error)/(auth_roles \[\]auth_role, err error)/' api/operations.go
	sed -i '' 's/(user \[\]user, err error)/(current_user \[\]user, err error)/' api/operations.go
# export enums that have a lowercase name
	sed -i '' 's/membership_roleReadOnlyUser/Membership_roleReadOnlyUser/' api/operations.go
	sed -i '' 's/membership_roleUser/Membership_roleUser/' api/operations.go
	sed -i '' 's/membership_roleTenantAdmin/Membership_roleTenantAdmin/' api/operations.go

## format
fmt:
	terraform fmt -recursive ./examples/
	go fmt ./...

## finds Go programs that use old APIs and rewrites them to use newer ones.
fix:
	go fix ./...

## run the lint aggreator golangci-lint over the codebase
lint:
	(which golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.2)
	$(GOBIN)/golangci-lint run ./...

## update go.mod to match the source code in the module
tidy:
	go mod tidy

## examines Go source code and reports suspicious constructs
vet:
	go vet ./...

## run tests
test:
	go test -cover ./...

## run acceptance tests against real infra
testacc:
	TF_ACC=1 go test -v ./...

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	TF_ACC=1 go test ./... -v -sweep -timeout 10m

# install into ~/go/bin, needed to generate the docs
install: $(GOBIN)/terraform-provider-prefect

$(GOBIN)/terraform-provider-prefect: internal/*/* api/* api/operations.go
	go install

## make docs
docs: $(GOBIN)/terraform-provider-prefect examples/*
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	touch docs

## apply examples
apply:
	terraform -chdir=examples apply

## show state for test project
state:
	terraform -chdir=examples state show prefect_project.test

## create ~/.terraformrc for local testing
~/.terraformrc: devtools/.terraformrc
	cat devtools/.terraformrc | sed "s|<GOBIN>|$(GOBIN)|" > ~/.terraformrc
