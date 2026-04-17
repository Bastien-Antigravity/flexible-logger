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
| **Notif Logger** | **Yes** | Real-time Apps | Fully Async | Capnp / Human | **Enabled** (Async) | **Local Reaction Queue** |

### Profile Details

*   **Audit (`NewAuditLogger`)**: Guaranteed delivery. The application waits until the network send is confirmed by the OS.
*   **Cloud Native (`NewCloudLogger`)**: Recommended for Kubernetes. Outputs structured JSON to stdout for easy collection by Fluentd/Datadog.
*   **Standard (`NewStandardLogger`)**: The most balanced profile. Local logs are readable and reliable, while network logs are handled in the background.
*   **High Performance (`NewHighPerfLogger`)**: Minimal overhead. Only sends logs over the network.
*   **Notif Logger (`NewNotifLogger`)**: Specialized for applications that need to react to errors programmatically. Provides a `SetLocalNotifQueue` method to pipe alerts into a Go channel.
*   **Developer (`NewDevelLogger`)**: Synchronous and verbose. Captures source information (file/line) for all log levels. Ideal for local development.
*   **Minimal (`NewMinimalLogger`)**: Lightweight asynchronous console logging. No network or file dependencies.
*   **No Lock (`NewNoLockLogger`)**: Optimized for highly concurrent systems. Uses lock-free paths and Cap'n Proto binary serialization for maximum throughput.

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

    // 2. Initialize Logger (False: Suppression of remote NOTIF connects)
    logger := profiles.NewHighPerfLogger("MyApp", config, false)
    defer logger.Close()

    // 3. Log with format strings
    logger.Info("Application started")
    logger.Debug("Processing request #%d from %s", 123, "user-abc")

    // 4. Notifications (Warnings/Errors trigger this automatically if configured)
    logger.Warning("High memory usage: %d%%", 87)
}
```

### Instantiation Examples

Each profile is optimized for a specific scenario:

```go
import "github.com/Bastien-Antigravity/flexible-logger/src/profiles"

// 1. Standard (Balanced)
logger := profiles.NewStandardLogger("my-service", config, false)

// 2. Cloud Native (JSON for Kubernetes)
logger := profiles.NewCloudLogger("my-service", config, false)

// 3. Audit (Strict Compliance - Blocking)
logger := profiles.NewAuditLogger("my-service", config, false)

// 4. High Perf (Network Only)
logger := profiles.NewHighPerfLogger("my-service", config, false)

// 5. No Lock (Binary Cap'n Proto)
logger := profiles.NewNoLockLogger("my-service", config, false)

// 6. Notif Logger (Local Event Reaction)
notifLogger := profiles.NewNotifLogger("my-service", config, true)
notifLogger.SetLocalNotifQueue(myNotifChannel) // React to errors in-process

// 7. Minimal (Simple Console)
logger := profiles.NewMinimalLogger("cli-app", false)

// 8. Development (Verbose & Synchronous)
logger := profiles.NewDevelLogger("dev-app", false)
```

```
