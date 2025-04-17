NAME=prefect
BINARY=terraform-provider-${NAME}

TESTS?=""
LOG_LEVEL?="INFO"
SWEEP?=""

default: build
.PHONY: default

help:
	@echo "Usage: $(MAKE) [target]"
	@echo ""
	@echo "This project defines the following build targets:"
	@echo ""
	@echo "  build            - compiles source code to build/"
	@echo "  clean            - removes built artifacts"
	@echo "  lint             - run static code analysis"
	@echo "  test             - run automated unit tests"
	@echo "  testacc          - run automated acceptance tests"
	@echo "  testacc-sweepers - run automated acceptance tests sweepers"
	@echo "  testacc-dev      - run automated acceptance tests from a local machine (args: TESTS=<tests or empty> LOG_LEVEL=<level> SWEEP=<yes or empty>)"
	@echo "  testacc-dev-user - run automated acceptance tests for user-related resources from a local machine"
	@echo "  testacc-oss      - run automated acceptance tests for Prefect OSS"
	@echo "  docs             - builds Terraform documentation"
	@echo "  dev-new          - creates a new dev testfile (args: resource=<resource> name=<name>)"
	@echo "  dev-clean        - cleans up dev directory"
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
	gotestsum --max-fails=50 ./... -run "$(TESTS)"

.PHONY: test

# NOTE: Acceptance Tests create real infrastructure
# against a dedicated testing account
testacc:
	TF_ACC=1 TESTS=$(TESTS) make test
.PHONY: testacc

# NOTE: Acceptance Test sweepers delete real infrastructure against a dedicated testing account
testacc-sweepers:
	go test ./internal/sweep -v -sweep=all
.PHONY: testacc-sweepers

testacc-dev:
	./scripts/testacc-dev $(TESTS) $(LOG_LEVEL) $(SWEEP)
.PHONY: testacc-dev

testacc-dev-user:
	./scripts/testacc-dev-user
.PHONY: testacc-dev-user

testacc-oss:
	docker compose up -d
	./scripts/testacc-oss $(TESTS)
.PHONY: testacc-oss

docs:
	rm -rf docs/ && mkdir docs/
	tfplugindocs generate --rendered-provider-name Prefect --provider-name prefect
.PHONY: docs

dev-new:
	sh ./create-dev-testfile.sh $(resource) $(name)
.PHONY: dev-new

dev-clean:
	rm -rf ./dev
.PHONY: dev-clean
