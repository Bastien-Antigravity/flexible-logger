GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: all build clean test lint

all: build

build:
	$(GOBUILD) ./...

clean:
	$(GOCLEAN)
	rm -rf bin/

test:
	$(GOTEST) -v ./...

lint:
	$(shell $(GOCMD) env GOPATH)/bin/golangci-lint run ./...
