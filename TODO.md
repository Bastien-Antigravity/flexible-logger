# TODO: flexible-logger

## 🚨 High Priority (Governance Gaps)
- [ ] **Level Purge (Purger Rule)**: Reduce the 12 log levels to 5 core levels (Debug, Info, Warn, Error, Critical). Use Tags for special categories (FEAT-001). (Approval Required)

## 🏗️ Architecture & Refactoring
- [ ] Standardize the Cap'n Proto schema for cross-service logging.

## 🧪 Testing & CI/CD
- [ ] Add benchmarks for `sync.Pool` performance under high load.

## ✅ Completed
- [x] **v1.3.4 Reliability Patch**:
    - Fixed `mapLevel` data loss in Cap'n Proto serializer (now supports all 12 levels).
    - Fixed memory leak in `MultiSink` with zero sinks (FEAT-004).
    - Fixed goroutine leak in `RemoteNotifier` tests.
- [x] **Test Suite Expansion**: Added ~20 new tests covering pooling, ref-counting, smart metadata, and concurrency stress.
- [x] **Audit Reliability**: Implemented `MultiSink` support for the Audit profile (Console + File + Blocking Network).
- [x] Initial BDD Spec migration.