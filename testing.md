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

### 6. Log Server Connectivity Tool
- **File**: `cmd/test-log-server/main.go`
- **Role**:
    - A standalone client used to verify connectivity to a real `log-server` at `127.0.0.2:9020`.
    - Includes `connection_test.go` for automated integration and benchmark testing against external servers.

### 7. Mock Infrastructure
- **File**: `src/test_utils/mock_server.go`
- **Role**:
    - **`StartMockServer`**: Provides a lightweight TCP sink that discards data. 
    - Used in benchmarks and integration tests to simulate Log and Notification servers without external dependencies.
    - Resolves "connection refused" errors in isolated environments like GitHub Actions.

## Running Tests

To execute the full test suite, use the following commands from the project root:

### Standard Unit Tests
```bash
make test
```

### Race Conditions Check
To ensure there are no concurrency issues in the asynchronous logging paths, the CI/CD pipeline runs tests with the race detector enabled:
```bash
go test -race ./...
```
All mock implementations (`MockSink`, `MockNotifier`) are thread-safe using `sync.Mutex`.

### Benchmarks & Integration Tests
The project includes several tools for performance and connectivity testing:

#### 1. Internal Benchmark
```bash
go run cmd/test/main.go
```
This test:
1.  Starts two local mock servers (Logs and Notifications).
2.  Initializes a `HighPerfLogger` pointing to these mocks.
3.  Logs 1,000,000 messages and measures throughput (logs/sec).

#### 2. Log Server Client & Benchmark
```bash
# Run standalone connectivity test
go run cmd/test-log-server/main.go

# Run connectivity unit test and benchmark
go test -v ./cmd/test-log-server/...
go test -v -bench=. ./cmd/test-log-server/...
```

## Diagnostics & Internal Reporting

The library includes a centralized internal error reporter in `src/error_handler/fallback_logger.go`. 

If an internal component fails (e.g., a background sink buffer overflows or a network connection is lost), the failure is reported to **`os.Stderr`** using the project's standard text format. This provides critical visibility during long-running integration tests or high-load benchmarks.

## Continuous Integration

The enhanced CI/CD pipeline (`.github/workflows/ci-cd.yml`) performs the following steps:
- **Linting**: Uses `golangci-lint` to maintain code quality.
- **Verbose Unit Testing**: Runs all tests with `-v` and `-race`.
- **Benchmark Verification**: Executes the 1,000,000 log benchmark to ensure stability under load and verify that no "connection refused" errors occur.
