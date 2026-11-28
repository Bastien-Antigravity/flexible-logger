---
microservice: flexible-logger
type: note
status: active
tags:
- '#service/flexible-logger'
- '#domain/observability'
- '#type/note'
- '#state/active'
- '#zone/3-fleet'
---

# TODO: flexible-logger

## 🧪 Pending Multi-Service Scenarios (Staged in `sandbox-testing`)
- [ ] **FEAT-023 End-to-End Scenarios (`sandbox-testing/02-Scenarios/go/logger_resilience_and_alerts_test.go`)**:
  - [ ] **Scenario 1**: Saturate `AsyncSink` buffer with 10k messages -> assert at most 1 warning alert dispatched within 60s throttle window with tag `#telegram` to mock `notif-server` with accurate drop count.
  - [ ] **Scenario 2**: Canonical SSoT Cap'n Proto zero-copy serialization over framed TCP socket -> deserialize and assert field fidelity in `log-server` mock sink.
  - [ ] **Scenario 3**: Initiate shutdown under active logging load -> assert `Close()` terminates within the bounded drain timeout (<= 2.5s) without process hangs or leaked goroutines.

## ✅ Completed
- [x] **RollingFileWriter Rename-Failure Safety & TextSerializer Newline Escaping (`src/sink/rolling_file.go`, `src/serializers/text_serializer.go`)**: Prevented `rotate()` from truncating active logs on `os.Rename` errors and escaped `\r`/`\n` in `TextSerializer` messages.
- [x] **LogEngine & AsyncSink Race Fixes + Stderr Storm Suppression (`src/engine/log_engine.go`, `src/sink/async.go`)**: Eliminated concurrent data races on `Level`, `CallerSkip`, `SamplingRate`, `OnError`, and drop-alert fields; suppressed redundant `os.Stderr` floods on `ErrBufferFull` and unbound `LocalNotifier`.
- [x] **Notifier Drop Accounting & Dead-Worker Close Guard (`src/notifier/local_notifier.go`, `remote_notifier.go`)**: Added `DroppedCount()` to both notifiers and marked `RemoteNotifier` closed if initial `worker()` connection initialization fails.
- [x] **Profile Consolidation, Nil-Config Safety & Audit TempDir Fallback (`src/profiles/common.go`, `audit.go`, `standard.go`, `cloud_native.go`, `high_perf.go`, `no_lock.go`, `notif_logger.go`, `devel.go`)**: Extracted shared `resolveServerCap`/`attachNotifier` helpers, fixed duplicate console fallback, moved audit emergency fallback to `os.TempDir()`, and wired `SetDropNotifier` across all async network profiles.
- [x] **Test Isolation & Hygiene (`src/profiles/profiles_test.go`, `src/helpers/paths_test.go`, `tests/integration_test.go`, `cmd/test-log-server/`)**: Isolated tests with `LOG_DIR` / `CONFIG_PATH` temp fixtures, expanded `TestAllProfiles` to all 8 profiles, and removed `AUDIT_EXTERNAL.md` and `cmd/*/standalone.yaml` symlinks.
- [x] **Canonical SSoT Cap'n Proto Schemas**: Centralized `logger.capnp` and `notifier.capnp` in `microservice-toolbox/go/pkg/schemas/{logger,notifier}` (`schemas/platform/`), migrated `flexible-logger`, `notif-server`, and `log-server` to SSoT, and deleted duplicate local schemas.
- [x] **Rate-Throttled Telegram Buffer-Drop Alerting**: Implemented atomic CAS 60s throttling window in `AsyncSink`, aggregate window drop calculation, and async warning dispatch with `Tags: []string{"telegram"}` via `dropNotifier.Notify()`. Wired in `NewStandardLogger` and `NewHighPerfLogger`.
- [x] **Bounded Shutdown Drain Timeout**: Replaced unbounded drain channel in `AsyncSink.Close()` with configurable 2.0s timeout (`drainTimeout`), forced downstream sink termination on timeout, and verified unblocking of worker goroutines.
- [x] **Resilient AuditLogger**: Implemented automatic fallback to `filepath.Join(os.TempDir(), "emergency_audit.log")` and `[LOGGER-AUDIT-ERROR]` diagnostic stderr logging if configured audit file creation fails.
- [x] **Audit Emergency Fallback & Wire Tests**: Added `TestAuditLogger_EmergencyFileFallback`, `TestAsyncSink_Close_DrainTimeout`, `TestAsyncSink_ThrottledDropAlert_PreventsTelegramStorm`, and updated `TestCapnpSerializer`.
- [x] **Unified Caller Stack Depth (`Log` vs. Helper Parity)**: Unified `LogEngine.Log()` and all severity helper methods (`Debug`, `Info`, `Warning`, `Error`, `Critical`, etc.) through `logInternal()` so `runtime.Caller(3 + l.getCallerSkip())` resolves identical caller `Filename` and `FunctionName` attribution regardless of entry point.
- [x] **Observable Async Drop Counter (`AsyncSink.DroppedCount()`)**: Added lock-free `droppedCount atomic.Uint64` counter and `DroppedCount() uint64` accessor on `AsyncSink` incremented whenever buffer saturation drops an entry.
- [x] **v0.0.1 Reliability Patch**:
    - Fixed `mapLevel` data loss in Cap'n Proto serializer (now supports all 12 levels).
    - Fixed memory leak in `MultiSink` with zero sinks (FEAT-004).
    - Fixed goroutine leak in `RemoteNotifier` tests.
- [x] **Test Suite Expansion**: Added ~20 new tests covering pooling, ref-counting, smart metadata, and concurrency stress.
- [x] **Audit Reliability**: Implemented `MultiSink` support for the Audit profile (Console + File + Blocking Network).
- [x] Initial BDD Spec migration.
- [x] Fixed `GetDefaultLogPath` missing helper in `src/helpers/paths.go`.
- [x] Aligned `src/serializers/serializer_test.go` with 8-column log-server format.
- [x] Cleansed tracked 8.0MB binary and 10 log files from git tree.
- [x] Standardized `json_serializer.go` to canonical Unix newline (`\n`) for container stream compliance.
- [x] Hardened `LocalNotifier` with `sync.RWMutex` for thread-safe dynamic queue binding.
- [x] Realigned all 6 BDD specs in `obsidian-brain` (`FEAT-001` through `FEAT-006`) to `microservice: flexible-logger` and `#domain/observability`.
- [x] Normalized AI documentation (`AI-Init.md`, `AGENTS.md`, and `Testing-Playbook.md`).