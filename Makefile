NAME=prefect
BINARY=terraform-provider-${NAME}

default: build
.PHONY: default

help:
	@echo "Usage: $(MAKE) [target]"
	@echo ""
	@echo "This project defines the following build targets:"
	@echo ""
	@echo "  build - compiles source code"
	@echo "  lint - run static code analysis"
	@echo "  test - run automated tests"
	@echo "  clean - removes built artifacts"
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
	gotestsum --max-fails=10 ./...
.PHONY: test
