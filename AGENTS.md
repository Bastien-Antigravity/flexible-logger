# AGENTS.md: flexible-logger

## Service Mission & Architecture Role
`flexible-logger` provides extensible logging adapters and formatting engines for the ecosystem, serving as a flexible sink provider for `universal-logger` and standalone components.

- **Ecosystem Role**: Logging subsystem and adapter layer.
- **Interfaces**: Compatible with `universal-logger` `ILogger` standard.
- **Configuration Link**: `standalone.yaml -> ../docker-deployment/modes/local/config/native.yaml`

## Key Build & Test Commands
```bash
# Run tests
go test -v ./...

# Build CLI/test runner
go build -o bin/flexible-logger ./cmd/flexible-logger
```

## AI Development & Integration Guidelines
1. **Interface Compliance**: Ensure logger structs fulfill the ecosystem `ILogger` contract (`Debug`, `Info`, `Warning`, `Error`, `Critical`).
2. **No Unbuffered File Leaks**: Always flush sinks on shutdown and close handlers properly.
3. **Header Ritual**: All source files MUST begin with the Triple-Block header (`ESSENTIAL PROCESS`, `DATA FLOW`, `KEY PARAMETERS`).
4. **Section Dividers**: Use `// -----------------------------------------------------------------------------` between exported methods.
