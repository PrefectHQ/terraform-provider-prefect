NAME=prefect
BINARY=terraform-provider-${NAME}

default: build test
.PHONY: default

help:
	@echo "Usage: $(MAKE) [target]"
	@echo ""
	@echo "This project defines the following build targets:"
	@echo ""
	@echo "  build - compiles source code"
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

test:
	gotestsum --max-fails=10 ./...
.PHONY: test
