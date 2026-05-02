# 🧬 Project DNA: flexible-logger

## 🎯 High-Level Intent (BDD)
- **Goal**: Provide a high-performance, flexible logging library for internal Go microservices with support for multiple backends.
- **Key Pattern**: **Adapter/Strategy Pattern** (pluggable backends: stdout, file, log-server) and **Zero-Allocation Logging** (using zeropool).
- **Behavioral Source of Truth**: [[business-bdd-brain/02-Behavior-Specs/flexible-logger]]
- **Spec Gate**: [HARDENED] No implementation without an `approved` spec in the folder above.

## 🛠️ Role Specifics
- **Architect**: 
    - Maintain zero-allocation performance on the hot paths.
    - Ensure that backend switching is transparent to the caller.
- **QA**: 
    - Verify that logs are correctly routed based on the active profile.
    - Test concurrent writes to the same backend to ensure thread safety.
- **Developer**:
    - Follow the established `Level` and `NotifMessage` mirroring patterns to maintain parity with `universal-logger`.

## 🚦 Lifecycle & Versioning
- **Primary Branch**: `develop`
- **Protected Branches**: `main`, `master`
- **Versioning Strategy**: Semantic Versioning (vX.Y.Z).
- **Version Source of Truth**: `VERSION.txt`.
