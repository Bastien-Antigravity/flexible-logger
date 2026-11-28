---
microservice: flexible-logger
type: general
status: active
tags:
- '#ai/ignore'
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
- '#type/general'
- '#state/active'
---
# General & Miscellaneous Guidelines

This document contains miscellaneous guidelines, fallback behaviors, and operational integration details for the Flexible Logger library.

---

## 🚨 Internal Diagnostics & Fallback Logger

To prevent infinite recursive loops (where logging an error with the logger causes another error, which attempts to log again), the library maintains a strict, isolated diagnostic recovery mechanism:

### 1. Fallback Path
*   **Source File**: `src/error_handler/fallback_logger.go`
*   **Behavior**: If any core logger component encounters a genuine infrastructure failure (e.g., lost network connections, disk write failures, or bad Cap'n Proto framing), the failure is routed away from standard sinks and written directly to **`os.Stderr`**. Expected non-blocking drop conditions (`sink.ErrBufferFull`) and unbound local notification queues are tracked via `DroppedCount()` and throttled drop alerts without flooding `os.Stderr` per message.
*   **Format**: Emitted as raw, human-readable text prefixed with `[LOGGER-INTERNAL-ERROR]` to ensure visibility without relying on network or file infrastructure.

### 2. Operational Impact
During high-load testing or staging environment networking failures, keep an eye on standard error output (`os.Stderr`). If a remote server falls behind or sinks encounter errors, diagnostic error notices are routed to standard error to prevent silent data loss.

---

## 🏷️ Ecosystem Tag Taxonomy

All observability documents and log-aggregation filters in the Bastien-Antigravity fleet rely on standard tags to organize indices and telemetry. When writing application configs or integrating the logger:

*   **Service Core Tag**: Always label the service using `#service/flexible-logger`.
*   **observability Domain**: Standardize under the `#domain/observability` umbrella.
*   **Ecosystem Tier**: Position all logging pipeline systems under `#zone/3-fleet`.

---

## 🛠️ Frequently Asked Questions (FAQ)

### How do I swap profiles dynamically?
All operational logger profiles are fully integrated with the `distributed-config` package. By updating the active configuration YAML (e.g., changing the logger profile key to `CloudNative` or `Audit` in the config server), the backend logger adjusts its sink topology on subsequent restarts.

### Can the logger block my application main thread?
* **Audit Profile**: Yes, by design. The Audit profile enforces synchronous local disk writes and blocking network TCP socket delivery (`ModeIndefinite`), waiting for socket transmission before returning.
* **Standard Profile**: Synchronous on local destinations (Console and local File) to guarantee zero local log loss, but asynchronous on network log streaming.
* **CloudNative / HighPerf / NoLock / Minimal**: Fully asynchronous. Application calls push to in-memory buffered channels (`AsyncSink`) and return immediately. If an asynchronous buffer fills to capacity under extreme pressure, entries are dropped according to the non-blocking `DropNewest` policy to prevent application deadlocks. (Probabilistic sampling is a distinct, configurable feature evaluated at ingestion).

### What happens during shutdown if a downstream network sink hangs?
`AsyncSink.Close()` implements a configurable **bounded drain timeout** (default 2.0s via `SetDrainTimeout`). If downstream network writers hang, the drain loop times out, forcefully closes downstream writers, and returns immediately so service shutdown is never blocked.

### How are buffer drops monitored without flooding Telegram or notif-server?
Drops are counted lock-free via `AsyncSink.DroppedCount() uint64`. Whenever drops occur, an atomic CAS rate-limiter ensures at most **one** warning notification is dispatched every 60 seconds (tagged `telegram`) with the count of drops in that window and cumulative drops.

### What happens if the configured audit log directory cannot be created or accessed?
`NewAuditLogger` catches the filesystem error, prints a loud `[LOGGER-AUDIT-ERROR]` diagnostic to `os.Stderr`, and immediately falls back to writing records to `filepath.Join(os.TempDir(), "emergency_audit.log")`.
