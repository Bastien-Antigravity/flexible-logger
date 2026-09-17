---
microservice: flexible-logger
type: rules
status: active
tags:
- '#service/flexible-logger'
- '#domain/observability'
- '#type/rules'
- '#state/active'
- '#zone/3-fleet'
---

# AGENTS.md: flexible-logger

## Service Mission & Architecture Role
`flexible-logger` provides extensible logging adapters and formatting engines for the ecosystem, serving as a flexible sink provider for `universal-logger` and standalone components.

- **Ecosystem Role**: Logging subsystem and adapter layer.
- **Interfaces**: Compatible with `universal-logger` `ILogger` standard.
- **Configuration Link**: `standalone.yaml -> ../docker-deployment/modes/local/config/native.yaml`

## Key Build & Test Commands
```bash
# Run unit tests
go test -v ./...

# Run race condition checks
go test -race ./...

# Build CLI demo / benchmark
go build -o bin/flexible-logger ./cmd/flexible-logger
```

## AI Development & Integration Guidelines
1. **Interface Compliance**: Ensure logger structs fulfill the ecosystem `ILogger` contract (`Debug`, `Info`, `Warning`, `Error`, `Critical`).
2. **Never Crash Host in Logger**: Never call `os.Exit(1)` inside logger constructors or sinks. Logging failures must fail gracefully to `os.Stderr` and fallback to local sinks.
3. **No Unbuffered File Leaks**: Always flush sinks on shutdown and close handlers properly.
4. **Rust Log-Server Format Parity**: `TextSerializer` must strictly adhere to the fixed-width 8-column layout (`timestamp`, `hostname`, `logger`, `level`, `filename`, `func`, `line`, `msg`, `[metadata: ...]`).
5. **Sink Reference Counting Contract**: In custom sinks, every `LogEntry` consumed must call `entry.Release()`. Async or fan-out sinks must retain references via `entry.Retain()`.
6. **Header Ritual**: All Go source files MUST begin with the Triple-Block header (`ESSENTIAL PROCESS`, `DATA FLOW`, `KEY PARAMETERS`).
7. **Section Dividers**: Use `// -----------------------------------------------------------------------------` between exported methods.
