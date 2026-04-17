# Testing Documentation

This document summarizes the testing infrastructure and coverage for the Flexible Logger library.

## Overview

The test suite is designed to ensure the reliability, performance, and thread-safety of the logging engine and its various components. It includes unit tests for core logic and integration tests for component interactions.

## Test Categories

## Testing Strategy

The `flexible-logger` uses a three-tier testing strategy to ensure high performance and reliability:

### 1. Unit Testing (Isolated)
We use the **Table-Driven Test** pattern to verify logical components in isolation without needing real IO.
*   **Profiles**: [profiles_test.go](file:///Users/imac/Desktop/Bastien-Antigravity/flexible-logger/src/profiles/profiles_test.go) automatically verifies that standalone profiles boot and log correctly in one loop.
*   **Engine**: [log_engine_test.go](file:///Users/imac/Desktop/Bastien-Antigravity/flexible-logger/src/engine/log_engine_test.go) uses **Mocks** to verify filtering, metadata, and notification triggers.
*   **Local Reaction**: [profiles_test.go](file:///Users/imac/Desktop/Bastien-Antigravity/flexible-logger/src/profiles/profiles_test.go) includes `TestNotifLogger_LocalQueue`, which verifies that the `NotifLogger` correctly pipes alerts (Warnings/Errors) into a Go channel for in-process handling.

### 2. Integration Testing (Cooperative)
These tests verify that the logger correctly interacts with the wider ecosystem.
*   **Config integration**: Ensuring that capabilities from `distributed-config` (like `log-server` IP/Port) are correctly mapped to sinks.
*   **Network Manager**: Verifying TCP connection, retry logic, and blocking/non-blocking behavior.

### 3. Verification & Benchmarking
*   **Performance**: Run `go test -v -bench=. ./...` to verify zero-allocation goals.
*   **Sanity Checks**: Use [verify_premium_features.go](file:///Users/imac/Desktop/Bastien-Antigravity/flexible-logger/scratch/verify_premium_features.go) for human-readable verification of JSON, Audit, and Sampling.

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

#### 3. Premium Features Verification
```bash
# Verify JSON, Audit, and Sampling behavior
go run scratch/verify_premium_features.go
```
This script validates that the `CloudNative` profile correctly produces JSON lines and that the `SamplingRate` logic effectively reduces log volume without dropping critical errors.

## Diagnostics & Internal Reporting

The library includes a centralized internal error reporter in `src/error_handler/fallback_logger.go`. 

If an internal component fails (e.g., a background sink buffer overflows or a network connection is lost), the failure is reported to **`os.Stderr`** using the project's standard text format. This provides critical visibility during long-running integration tests or high-load benchmarks.

## Continuous Integration

The enhanced CI/CD pipeline (`.github/workflows/ci-cd.yml`) performs the following steps:
- **Linting**: Uses `golangci-lint` to maintain code quality.
- **Verbose Unit Testing**: Runs all tests with `-v` and `-race`.
- **Benchmark Verification**: Executes the 1,000,000 log benchmark to ensure stability under load and verify that no "connection refused" errors occur.
