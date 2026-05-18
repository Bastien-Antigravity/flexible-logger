---
microservice: flexible-logger
type: general
status: active
tags:
- '#ai/ignore'
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
---

# General & Miscellaneous Guidelines

This document contains miscellaneous guidelines, fallback behaviors, and operational integration details for the Flexible Logger library.

---

## 🚨 Internal Diagnostics & Fallback Logger

To prevent infinite recursive loops (where logging an error with the logger causes another error, which attempts to log again), the library maintains a strict, isolated diagnostic recovery mechanism:

### 1. Fallback Path
*   **Source File**: `src/error_handler/fallback_logger.go`
*   **Behavior**: If any core logger component encounters a severe infrastructure failure (e.g., asynchronous ring buffer overflows, lost network connections, or bad Cap'n Proto framing), the failure is routed away from standard sinks and written directly to **`os.Stderr`**.
*   **Format**: Emitted as raw, human-readable text prefixed with `[LOGGER-INTERNAL-ERROR]` to ensure visibility without relying on network or file infrastructure.

### 2. Operational Impact
During high-load testing or staging environment networking failures, keep an eye on standard error output (`os.Stderr`). If a remote server falls behind, standard buffers drop performance-tier logs but write a single fallback drop alert to prevent silent data loss.

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
Only if you explicitly choose the **Audit** logger profile. The Audit profile implements a zero-drop blocking path. All other operational profiles (e.g., Standard, HighPerf, CloudNative) route standard statements via non-blocking asynchronous `AsyncSink` channels. If a sink buffer fills to maximum capacity, non-critical logs are probabilistically dropped, ensuring that your application's primary throughput never stalls due to network logging bottlenecks.
