---
microservice: flexible-logger
type: repository
status: active
language: go
tags:
  - domain/observability
---

# Flexible Logger

A high-performance, zero-allocation, asynchronous logging library for Go, designed for distributed systems.

## Features

*   **Zero Allocation**: Uses `sync.Pool` for log entries to minimize GC pressure.
*   **Asynchronous IO**: Non-blocking logging using buffered channels (`AsyncSink`).
*   **Structured & Binary**: Native support for **Cap'n Proto** serialization.
*   **Network Logging**: Reliable TCP logging with auto-reconnection (`NetworkManager`).
*   **Notifications**: Asynchronous alert system (`RemoteNotifier`) for warnings and errors.
*   **Automatic Metadata**: Captures `ProcessID`, `Hostname`, `Filename`, and `LineNumber`.
*   **Smart Sampling**: Probabilistic log dropping for high-traffic (never drops Errors).
*   **Audit Trail**: Zero-drop blocking mode for critical compliance logs.
*   **Flexible Config**: Hot-swappable configurations via `distributed-config`.

## Logger Profiles

The library provides several pre-configured profiles tailored for different environments.

| Profile | Distributed Config | Target Use Case | IO Behavior | Format (File) | Network Logs | **Log Drop Policy** |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Developer** | **No** | Local Coding | Synchronous | Human Text | Disabled | **Never Drops** |
| **Standard** | **Yes** | Staging / Prod | Mixed | Human Text | **Enabled** (Async) | **Zero Local Loss** (Drops Net) |
| **Cloud Native**| **Yes** | Kubernetes/Cloud | Fully Async | **JSON** | **Enabled** (Async) | **Best Effort** (Drops on Saturation) |
| **Audit** | **Yes** | Compliance/Legal | **Blocking** | Human Text | **Enabled** (**Blocking**) | **Never Drops (Secure)** |
| **High Perf** | **Yes** | High Load | Async Net | *None* | **Enabled** (Async) | **Best Effort** (Drops on Saturation) |
| **No Lock** | **Yes** | Concurrency | Fully Async | Binary Capnp | **Enabled** (Async) | **Best Effort** (Drops All) |
| **Minimal** | **No** | Simple CLIs | Async | *None* | Disabled | **Best Effort** (Drops Console) |

### Profile Details

*   **Audit (`NewAuditLogger`)**: Guaranteed delivery. The application waits until the network send is confirmed by the OS.
*   **Cloud Native (`NewCloudLogger`)**: Recommended for Kubernetes. Outputs structured JSON to stdout for easy collection by Fluentd/Datadog.
*   **Standard (`NewStandardLogger`)**: The most balanced profile. Local logs are readable and reliable, while network logs are handled in the background.
*   **High Performance (`NewHighPerfLogger`)**: Minimal overhead. Only sends logs over the network.

## Metadata & Performance

The logger automatically enriches every log entry with system metadata. To balance detail with extreme performance, we use a **Smart Collection** policy:

*   **Static Metadata (Always On)**: `ProcessID`, `ProcessName`, and `Hostname` are cached at startup and added to every log with zero performance impact.
*   **Dynamic Metadata (Selective)**: Caller information (`Filename`, `LineNumber`, `FunctionName`) is captured using `runtime.Caller`.
    *   **HighPerf/Standard Profiles**: Only captures source info for **Warning**, **Error** and **Critical** logs. Standard `Info` logs remain lightning fast.
    *   **Development Profile**: Captures source info for **all** log levels to aid debugging.

## Usage

All log methods use **Printf-style** format strings (`format string, args ...any`), so you can embed variables directly without `fmt.Sprintf`:

```go
package main

import (
    "github.com/Bastien-Antigravity/flexible-logger/src/profiles"
    distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

func main() {
    // 1. Load Config
    config := distributed_config.New("standalone")

    // 2. Initialize Logger
    logger := profiles.NewHighPerfLogger("MyApp", config)
    defer logger.Close()

    // 3. Log with format strings
    logger.Info("Application started")
    logger.Debug("Processing request #%d from %s", 123, "user-abc")

    // 4. Notifications (Warnings/Errors trigger this automatically if configured)
    logger.Warning("High memory usage: %d%%", 87)
}
```

```
