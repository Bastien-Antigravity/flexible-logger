# TODO: flexible-logger

## 🚨 High Priority (Governance Gaps)
- [ ] **Level Purge (Purger Rule)**: Reduce the 12 log levels to 5 core levels (Debug, Info, Warn, Error, Critical). Use Tags for special categories (FEAT-001). (Approval Required)
- [ ] **Pool Safety**: Ensure `LogEntry.Release()` is called even if `Sink.Write` returns an error to prevent memory leaks under failure conditions (FEAT-004). (Approval Required)

## 🏗️ Architecture & Refactoring
- [ ] Implement `MultiSink` support for the Audit profile.
- [ ] Standardize the Cap'n Proto schema for cross-service logging.

## 🧪 Testing & CI/CD
- [ ] Add benchmarks for `sync.Pool` performance under high load.

## ✅ Completed
- [x] Initial BDD Spec migration.