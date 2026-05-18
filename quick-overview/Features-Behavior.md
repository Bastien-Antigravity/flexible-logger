---
microservice: flexible-logger
type: behavior
status: active
tags:
- '#ai/ignore'
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
---

# Features & Behavior

This document details the operational behavior, profiles, metadata policies, and connection resilience strategies of the Flexible Logger.

---

## Core Features

*   **Zero Allocation**: Employs a robust `sync.Pool` for `LogEntry` structures to dramatically minimize Garbage Collection (GC) pressure and overhead on performance-critical paths.
*   **Asynchronous IO**: Decouples application execution context from I/O wait times using highly-optimized background channels via `AsyncSink`.
*   **Structured & Binary Formats**: Implements native Cap'n Proto serialization (supporting full 12-level ecosystem parity) as well as standard structured JSON for Kubernetes collection.
*   **Network Resilience**: Integrates directly with the ecosystem's `conn_manager` connection resilience strategies (Critical, Standard, Performance) to ensure robust TCP communication.
*   **Decoupled Alerts & Notifications**: High-severity errors and warnings bypass standard logging pipelines, flowing through the asynchronous `RemoteNotifier` or custom programmatic `LocalNotifier` queues.
*   **Smart Metadata Enrichment**: Cached static fields (like `Hostname` and `ProcessID`) are embedded at zero cost, while caller details (file/line) are dynamically collected based on active log levels.
*   **Probabilistic Log Sampling**: Implements smart sampling to handle high-traffic environments gracefully, ensuring non-critical lines are dropped safely while guaranteed delivery is enforced for all errors.
*   **Real-Time Inspection**: Exposes the `GetLevel()` accessor, allowing real-time audit tools and test runner processes to examine active log levels dynamically.

---

## Ecosystem-Wide Logger Profiles

The logger provides pre-packaged profiles engineered for specific operational environments.

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

---

## 🛡️ Connection Resilience Strategies

`flexible-logger` delegates underlying connection and recovery mechanisms to standard `microservice-toolbox` presets:

*   **Critical Strategy (Audit Profile)**: Implements **Indefinite Retry** mode with exponential backoff and randomized jitter. It guarantees that compliant audit trails are maintained, blocking standard processing if the network endpoint becomes unreachable until a connection is restored.
*   **Standard Strategy (Default)**: Combines robust backoff rules with a maximum recovery wait-cap (2.0s), striking a clean balance between networking reliability and overall system responsiveness.
*   **Performance Strategy (High-Perf)**: Optimized strictly for high-throughput, low-latency applications, using aggressive timeouts and fail-fast policies.

---

## Profile Instantiation Examples

Each logger profile is instantiated and optimized for specific runtime semantics:

```go
import (
    distributed_config "github.com/Bastien-Antigravity/distributed-config"
    "github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

// 1. Standard Profile (Balanced for standard production/staging environments)
logger := profiles.NewStandardLogger("my-service", config, false)

// 2. Cloud Native Profile (Emits structured JSON to stdout for container collectors)
logger := profiles.NewCloudLogger("my-service", config, false)

// 3. Audit Profile (Strict legal compliance - enforces synchronous blocking network delivery)
logger := profiles.NewAuditLogger("my-service", config, false)

// 4. High Performance Profile (Low overhead - network delivery only, skips local files)
logger := profiles.NewHighPerfLogger("my-service", config, false)

// 5. No Lock Profile (Designed for extreme concurrency, lock-free pathways, Cap'n Proto binary formats)
logger := profiles.NewNoLockLogger("my-service", config, false)

// 6. Notif Logger (Enables real-time error reactions via custom programmatic channels)
notifLogger := profiles.NewNotifLogger("my-service", config, true)
notifLogger.SetLocalNotifQueue(myNotifChannel) // Pipes alerts in-process

// 7. Minimal Profile (Lightweight console logging for basic command-line applications)
logger := profiles.NewMinimalLogger("cli-app", false)

// 8. Development Profile (Verbose, synchronous, and collects caller information on all levels)
logger := profiles.NewDevelLogger("dev-app", false)
```

---

## Metadata Collection & Sampling Policy

To achieve high throughput while capturing granular logs when they matter most, the logger operates a hybrid metadata collection policy:

*   **Static Metadata**: Standard fields like `ProcessID`, `ProcessName`, and `Hostname` are resolved at boot, cached in the logger context, and appended to log entries with zero runtime memory allocation.
*   **Dynamic Metadata**: Details like source files, line numbers, and active functions require calling Go's `runtime.Caller`.
    *   **Developer Profile**: Resolves caller context on **all** log levels to simplify local code trace operations.
    *   **Standard/High-Perf Profiles**: Selectively captures caller details for **Warning**, **Error**, and **Critical** events. Standard operational logs (`Info`, `Debug`) remain highly optimized.
*   **Remote Handshake Identification**: During connection initialization with Central Log and Notification servers, the logger explicitly transmits its program identity. This ensures that upstream aggregators can immediately map incoming socket streams to the exact service origin.

---

## Operational Safety and Concurrency Hardening

*   **Pool Lifecycle Safety**: Verified that every `LogEntry` retrieved from the pool has a guaranteed path to `Release()`. This lifecycle holds true across fan-out sinks (like `MultiSink`) and drop pipelines (like `AsyncSink`).
*   **Thread Safety**: Core interfaces and mocks are protected against concurrent access utilizing thread-safe primitives.
*   **Cap'n Proto Integrity**: Enforces strict, zero-loss serialization round-trips for all 12 operational log levels, preventing protocol decoding failures at the network gateway.
