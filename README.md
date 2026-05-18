---
microservice: flexible-logger
type: repository
status: active
language: go
version: 0.0.1
tags:
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
---

# Flexible Logger

![Version](https://img.shields.io/badge/version-0.0.1-blue)
![Stability](https://img.shields.io/badge/stability-production--ready-green)

A high-performance, zero-allocation, asynchronous logging library for Go, designed for distributed systems. 

**v0.0.1 (Initial Release)**: Features hardened memory management, full Cap'n Proto level parity, and verified concurrency safety.

## 📖 Documentation Directory Map

This repository separates main code entry points from rich human-onboarding and operational manuals. For deep-dives into logger behaviors, testing routines, and configurations, see the files inside `quick-overview/`:

*   [**Architecture Overview**](quick-overview/Architecture-Overview.md): Component diagrams (LogEngine, Sinks, Notifiers), internal `sync.Pool` pooling, and centralized connection mapping from the toolbox.
*   [**Features & Behavior**](quick-overview/Features-Behavior.md): Extensive detail on the 8 pre-configured **Logger Profiles** (Standard, Audit, Cloud Native, etc.), standardized connection resilience strategies (Critical, Standard, Performance), and smart dynamic metadata/caller sampling policies.
*   [**Testing Playbook**](quick-overview/Testing-Playbook.md): Strategy layers for unit, integration, and throughput benchmark execution. Guides for local mock TCP server orchestration and premium features validation.
*   [**General & Misc**](quick-overview/General-Misc.md): Safe fallback diagnostics via `os.Stderr` and tag taxonomies.

---

## Usage

All log methods use **Printf-style** format strings (`format string, args ...any`), so you can embed variables directly without `fmt.Sprintf`:

```go
package main

import (
	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	flexible_logger_profiles "github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

func main() {
	// 1. Load Config
	config := distributed_config.New("standalone")

	// 2. Initialize Logger (False: Suppression of remote NOTIF connects)
	logger := flexible_logger_profiles.NewHighPerfLogger("MyApp", config, false)
    defer logger.Close()

    // 3. Log with format strings
    logger.Info("Application started")
    logger.Debug("Processing request #%d from %s", 123, "user-abc")

    // 4. Notifications (Warnings/Errors trigger this automatically if configured)
    logger.Warning("High memory usage: %d%%", 87)
}
```

### Instantiation Examples

Each profile is optimized for a specific scenario:

```go
import "github.com/Bastien-Antigravity/flexible-logger/src/profiles"

// 1. Standard (Balanced)
logger := profiles.NewStandardLogger("my-service", config, false)

// 2. Cloud Native (JSON for Kubernetes)
logger := profiles.NewCloudLogger("my-service", config, false)

// 3. Audit (Strict Compliance - Blocking)
logger := profiles.NewAuditLogger("my-service", config, false)

// 4. High Perf (Network Only)
logger := profiles.NewHighPerfLogger("my-service", config, false)

// 5. No Lock (Binary Cap'n Proto)
logger := profiles.NewNoLockLogger("my-service", config, false)

// 6. Notif Logger (Local Event Reaction)
notifLogger := profiles.NewNotifLogger("my-service", config, true)
notifLogger.SetLocalNotifQueue(myNotifChannel) // React to errors in-process

// 7. Minimal (Simple Console)
logger := profiles.NewMinimalLogger("cli-app", false)

// 8. Development (Verbose & Synchronous)
logger := profiles.NewDevelLogger("dev-app", false)
```

```
