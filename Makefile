NAME=prefect
BINARY=terraform-provider-${NAME}

default: build
.PHONY: default

help:
	@echo "Usage: $(MAKE) [target]"
	@echo ""
	@echo "This project defines the following build targets:"
	@echo ""
	@echo "  build - compiles source code to build/"
	@echo "  clean - removes built artifacts"
	@echo "  lint - run static code analysis"
	@echo "  test - run automated unit tests"
	@echo "  testacc - run automated acceptance tests"
	@echo "  docs - builds Terraform documentation"
.PHONY: help

build: $(BINARY)
.PHONY: build

$(BINARY):
	mkdir -p build/
	go build -o build/$(BINARY)
.PHONY: $(BINARY)

clean:
	rm -vrf build/
.PHONY: clean

lint:
	golangci-lint run
.PHONY: lint

test:
	gotestsum --max-fails=50 ./...
.PHONY: test

# NOTE: Acceptance Tests create real infrastructure
# against a dedicated testing account
testacc:
	TF_ACC=1 make test
.PHONY: testacc

docs:
	mkdir -p docs
	rm -rf ./docs/images
	go generate ./...
.PHONY: docs
