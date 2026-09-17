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

## 🏗️ Architecture & Refactoring
- [ ] **Runtime Caller Skip Parameterization**: Currently, `LogEngine.getEntry()` uses a base skip offset of `3 + l.CallerSkip` (`runtime.Caller(3 + l.CallerSkip)`). This offset is currently calibrated for the wrapping facade in `universal-logger` (and helper methods like `Info`/`Error`). However, when `logger.Log(level, format, ...)` is called directly, the stack frame is 1 level shallower. Future refactoring should pass an explicit frame offset parameter rather than relying on a fixed base level, ensuring exact caller resolution whether called directly or via multi-layer library facades.
- [ ] Standardize the Cap'n Proto schema for cross-service logging.

## 🧪 Testing & CI/CD
- [ ] Add benchmarks for `sync.Pool` performance under high load.

## ✅ Completed
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