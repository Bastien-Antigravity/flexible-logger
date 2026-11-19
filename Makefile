VERSION := $(strip $(shell cat VERSION.txt))
$(if $(VERSION),,$(error VERSION.txt is missing or empty))

.PHONY: all build test version clean

all: build

version:
	@echo $(VERSION)

build:
	@echo "Building repository (version $(VERSION))..."
	@if [ -f "go.mod" ]; then go build ./...; fi

test:
	@echo "Running tests (version $(VERSION))..."
	@if [ -f "go.mod" ]; then go test -v ./...; fi

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dist build bin/ logs/ *.egg-info target/
