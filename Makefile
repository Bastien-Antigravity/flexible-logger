GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

CAPI_SRC=python/capi/main.go
PYTHON_LIB_DIR=python/flexible_logger
LIB_NAME=libflexible_logger

# Detect OS
ifeq ($(OS),Windows_NT)
    LIB_EXT=.dll
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        LIB_EXT=.so
    endif
    ifeq ($(UNAME_S),Darwin)
        LIB_EXT=.dylib
    endif
endif

LIB_OUT=$(PYTHON_LIB_DIR)/$(LIB_NAME)$(LIB_EXT)

.PHONY: all build clean test python-build build-capi

all: build

build: build-capi

build-capi:
	$(GOBUILD) -o $(LIB_OUT) -buildmode=c-shared $(CAPI_SRC)

python-build: build-capi
	cd python && python3 -m build

clean:
	rm -f $(PYTHON_LIB_DIR)/$(LIB_NAME).*
	rm -f $(PYTHON_LIB_DIR)/*.h

test:
	$(GOTEST) -v ./...
	if [ -d "python/tests" ]; then pytest python/tests; fi
