GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: all build clean test

all: build

build:
	$(GOBUILD) ./...

clean:
	$(GOCLEAN)
	rm -rf bin/

test:
	$(GOTEST) -v ./...
