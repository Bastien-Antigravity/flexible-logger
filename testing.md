# Testing Documentation

This document summarizes the testing infrastructure and coverage for the Flexible Logger library.

## Overview

The test suite is designed to ensure the reliability, performance, and thread-safety of the logging engine and its various components. It includes unit tests for core logic and integration tests for component interactions.

## Test Categories

### 1. Core Logging Engine
- **File**: `src/engine/log_engine_test.go`
- **Coverage**: 
    - Log level filtering (ensuring only logs above the threshold are processed).
    - Entry retrieval and recycling via `sync.Pool`.
    - Correct dispatching of log entries to sinks.
    - Notification triggers for high-priority levels (Warning and Error).

### 2. Sink Implementations
- **File**: `src/sink/sink_test.go`
- **Coverage**:
    - **`WriterSink`**: Verifies serialization and writing to an `io.Writer`.
    - **`AsyncSink`**: Tests asynchronous processing using buffered channels and worker goroutines.
    - **`MultiSink`**: Ensures log entries are correctly fanned out to multiple destinations.

### 3. Notification System
- **File**: `src/notifier/notifier_test.go`
- **Coverage**:
    - **`LocalNotifier`**: Verifies channel-based notification delivery.
    - **`RemoteNotifier`**: Tests basic message queuing for remote delivery.

### 4. Logger Profiles
- **File**: `src/profiles/profiles_test.go`
- **Coverage**:
    - Initialization of `Minimal` and `Devel` logger profiles.
    - Basic verification that initialized loggers can write messages without errors.

### 5. Network Management
- **File**: `src/network_manager/network_manager_test.go`
- **Coverage**:
    - Basic initialization and configuration of the `NetworkManager`.

## Running Tests

To execute the full test suite, use the following commands from the project root:

### Standard Unit Tests
```bash
make test
```

### Race Conditions Check
To ensure there are no concurrency issues in the asynchronous logging paths:
```bash
go test -race ./...
```

### Verbose Output
```bash
go test -v ./...
```

## Continuous Integration

All tests are automatically executed via GitHub Actions on every push to the `main` and `develop` branches. The configuration can be found in `.github/workflows/ci-cd.yml`.
