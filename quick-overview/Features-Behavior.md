---
microservice: flexible-logger
type: behavior
status: active
tags:
- '#ai/ignore'
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
- '#type/behavior'
- '#state/active'
---
# Features & Behavior

This document details the operational behavior, profiles, metadata policies, and connection resilience strategies of the Flexible Logger.

---

## Core Features

*   **Low Allocation**: Employs a robust `sync.Pool` for `LogEntry` structures to dramatically minimize Garbage Collection (GC) pressure and overhead on performance-critical paths.
*   **Asynchronous IO**: Decouples application execution context from I/O wait times using highly-optimized background channels via `AsyncSink`.
*   **Structured & Binary Formats**: Implements native Cap'n Proto serialization (supporting full 12-level ecosystem parity) as well as standard structured JSON for Kubernetes collection.
*   **Network Resilience**: Integrates directly with the ecosystem's `conn_manager` connection resilience strategies (Critical, Standard, Performance) to ensure robust TCP communication.
*   **Decoupled Alerts & Notifications**: High-severity errors and warnings bypass standard logging pipelines, flowing through the asynchronous `RemoteNotifier` or custom programmatic `LocalNotifier` queues.
*   **Throttled Buffer-Drop Alerting**: When async buffers saturate under log bursts (`DropNewest` policy), drops increment atomic drop counters (`DroppedCount()`) and emit throttled warning notifications (default 60s window) with `#telegram` tags to avoid overwhelming `notif-server` or Telegram operators.
*   **Bounded Graceful Shutdown**: `AsyncSink.Close()` guarantees termination within a bounded drain timeout (default 2s), preventing microservices from hanging during shutdown if downstream network sinks stall.
*   **Single Source of Truth (SSoT) Wire Schemas**: Cap'n Proto wire schemas (`logger.capnp`, `notifier.capnp`) are consumed directly from `microservice-toolbox/go/pkg/schemas/{logger,notifier}`, guaranteeing byte-level cross-service and cross-language parity.
*   **Resilient Audit Logging**: `NewAuditLogger` enforces synchronous dual-destination logging with an emergency fallback file (`emergency_audit.log`) in temporary storage and loud stderr diagnostics if the target log directory cannot be opened.
*   **Smart Metadata Enrichment**: Cached static fields (like `Hostname` and `ProcessID`) are embedded at zero cost, while caller details (file/line) are dynamically collected based on active log levels.
*   **Probabilistic Log Sampling**: Implements smart sampling to handle high-traffic environments gracefully, ensuring non-critical lines are dropped safely while guaranteed delivery is enforced for all errors. Supports atomic runtime updates via `SetSamplingRate(rate float64)`.
*   **Real-Time Inspection & Control**: Exposes thread-safe `GetLevel()` access and atomic/synchronized runtime updates via `SetLevel(level)`, `SetCallerSkip(skip)`, and `SetSamplingRate(rate)`.

---

## Ecosystem-Wide Logger Profiles

The logger provides pre-packaged profiles engineered for specific operational environments.

| Profile | Distributed Config | Target Use Case | IO Behavior | Format (File) | Network Logs | **Log Drop Policy** |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Developer** | **No** | Local Coding | Synchronous | Human Text | Disabled | **Never Drops** |
| **Standard** | **Yes** | Staging / Prod | Mixed (Sync File/Console + Async Net) | Human Text | **Enabled** (Async) | **Zero Local Loss** (Drops Net on Saturation) |
| **Cloud Native**| **Yes** | Kubernetes/Cloud | Fully Async | **JSON** | **Enabled** (Async) | **Best Effort** (Drops on Saturation) |
| **Audit** | **Yes** | Compliance/Legal | **Blocking** | Human Text | **Enabled** (**Blocking**) | **Never Drops Locally / Blocks Socket Write** |
| **High Perf** | **Yes** | High Load | Async Net | *None* | **Enabled** (Async) | **Best Effort** (Drops on Saturation) |
| **No Lock** | **Yes** | Concurrency | Fully Async | Human Text | **Enabled** (Async) | **Best Effort** (Drops All) |
| **Minimal** | **No** | Simple CLIs | Async | *None* | Disabled | **Best Effort** (Drops Console) |
| **Notif Logger** | **Yes** | Real-time Apps | Fully Async | Human Text | **Enabled** (Async) | **Local Reaction Queue** |

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

// 3. Audit Profile (Strict legal compliance - enforces synchronous local file writes and blocking network socket delivery)
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
    *   **Standard/High-Perf Profiles**: Selectively captures caller details for **Error** and **Critical** events. Standard operational logs (`Info`, `Debug`, `Warning`) remain highly optimized.
*   **Remote Handshake Identification**: During connection initialization with Central Log and Notification servers, the logger explicitly transmits its program identity. This ensures that upstream aggregators can immediately map incoming socket streams to the exact service origin.

---

## Operational Safety and Concurrency Hardening

*   **Pool Lifecycle Safety**: Verified that every `LogEntry` retrieved from the pool has a guaranteed path to `Release()`. This lifecycle holds true across fan-out sinks (like `MultiSink`) and drop pipelines (like `AsyncSink`).
*   **Thread Safety**: Core interfaces and mocks are protected against concurrent access utilizing thread-safe primitives.
*   **Cap'n Proto Integrity**: Enforces strict, zero-loss serialization round-trips for all 12 operational log levels, preventing protocol decoding failures at the network gateway.

---

## 🚨 Throttled Buffer-Drop Alerting & Telegram Notification

When asynchronous buffer queues (`AsyncSink`) saturate under severe logging spikes or network stalls, the logger enforces a non-blocking `DropNewest` policy:

1.  **Drop Counter Accounting**: Every dropped entry releases memory back to `sync.Pool` (`entry.Release()`) and increments an atomic drop counter (`s.droppedCount.Add(1)`). Callers can inspect cumulative drops at any time via `AsyncSink.DroppedCount() uint64`, as well as `LocalNotifier.DroppedCount()` and `RemoteNotifier.DroppedCount()` for notification queue saturation.
2.  **Atomic CAS Rate Throttling**: To prevent buffer overflow bursts from triggering alert storms (e.g. hundreds of thousands of Telegram messages or `notif-server` floods), alerting is gated by an atomic compare-and-swap timestamp check (`lastAlertNano`).
    *   **Default Window**: 60 seconds (`dropAlertInterval`). Configurable via `SetDropAlertInterval(d time.Duration)`.
    *   **Aggregation**: Computes total drops within the elapsed window (`droppedInWindow = totalDropped - s.lastReportedDrop`).
3.  **Telegram Alert Dispatch**: If a `dropNotifier` is wired (automatic across async profiles via `attachNotifier`: `NewStandardLogger`, `NewCloudLogger`, `NewHighPerfLogger`, `NewNoLockLogger`, and `NewNotifLogger`), an alert notification is dispatched asynchronously:
    *   **Tags**: `[]string{"telegram"}` (routing directly to Telegram alert channels).
    *   **Severity**: `models.LevelWarning`.
    *   **Payload**: `⚠️ [<service>] AsyncSink buffer saturated: <N> log(s) dropped in window (total dropped: <M>)`.
4.  **Observer Callback**: Applications can attach custom metrics callbacks via `SetOnDropAlert(cb func(droppedInWindow, totalDropped uint64))` to pipe drop metrics to Prometheus or StatsD.

---

## ⏱️ Graceful Shutdown & Bounded Drain Timeout

During microservice shutdown, unbuffered log drains must never block service termination indefinitely:

1.  **Drain Loop**: `AsyncSink.Close()` closes the entry channel and attempts to drain pending entries to the downstream sink via a completion channel (`drained := make(chan struct{})`).
2.  **Configurable Timeout**: By default, `drainTimeout` is set to **2.0 seconds** (`SetDrainTimeout(d time.Duration)`).
3.  **Forced Termination on Stalls**: If the downstream sink blocks (e.g., hanging TCP socket), the timeout branch terminates the drain loop, forces downstream closure (`s.next.Close()`), and returns an error formatted as `async sink close timeout after <duration>, forced shutdown`. This guarantees the background worker unblocks and the host process can terminate cleanly.

---

## 📜 Canonical SSoT Wire Schemas

All wire protocol schemas (`logger.capnp` and `notifier.capnp`) follow the ecosystem Single Source of Truth (SSoT) architecture:

*   **Canonical Source**: Located centrally in `microservice-toolbox/go/pkg/schemas/{logger,notifier}` and symlinked at `microservice-toolbox/schemas/platform/`.
*   **Zero Local Duplication**: `flexible-logger` imports `github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/schemas/logger` and `notifier` directly. No local duplicate copies of `.capnp` or generated Go files exist in the repository.
*   **Compliance Rule**: Verified continuously by the fleet compliance auditor (`validate_compliance.py`) with rules `DUPLICATE_PLATFORM_SCHEMA` and `LEGACY_LOCAL_SCHEMA_IMPORT`.

---

## 🛡️ Audit Logger Resilience & Fallback

The `Audit` profile (`NewAuditLogger`) provides strict legal compliance and non-repudiation:

*   **Blocking Semantics**: Enforces synchronous local file writes and blocking TCP socket delivery (`ModeIndefinite` retry strategy).
*   **Emergency Fallback File**: If opening the primary audit log file fails (e.g. missing target directory, permission denial), it logs a diagnostic message `[LOGGER-AUDIT-ERROR]` to `os.Stderr` and automatically falls back to an emergency file in temporary storage (`filepath.Join(os.TempDir(), "emergency_audit.log")`).
*   **Dual Destinations**: Binds both console, audit file, and blocking remote network sink via `MultiSink`, ensuring audit records are preserved even under network partition.
