# Flexible Logger

A high-performance, zero-allocation, asynchronous logging library for Go, designed for distributed systems.

## Features

*   **Zero Allocation**: Uses `sync.Pool` for log entries to minimize GC pressure.
*   **Asynchronous IO**: Non-blocking logging using buffered channels (`AsyncSink`).
*   **Structured & Binary**: Native support for **Cap'n Proto** serialization.
*   **Network Logging**: Reliable TCP logging with auto-reconnection (`NetworkManager`).
*   **Notifications**: Asynchronous alert system (`RemoteNotifier`) for warnings and errors.
*   **Flexible Config**: Hot-swappable configurations via `distributed-config`.

## Profiles

The library provides pre-configured profiles for common use cases:

### 1. High Performance (`NewHighPerfLogger`)
*   **Target**: Production, High-Load.
*   **Behavior**: Fully Asynchronous.
*   **Sinks**: Network Sink (Async).
*   **Reliability**: Drops logs if buffer is full (favors application performance over log completeness).

### 2. Standard (`NewStandardLogger`)
*   **Target**: General Production / Staging.
*   **Behavior**: Mixed Sync/Async.
*   **Sinks**: Console (Sync), File (Sync), Network (Async).
*   **Reliability**: Blocks on File/Console writes to ensure local persistence.

### 3. No Lock (`NewNoLockLogger`)
*   **Target**: Extreme Concurrency.
*   **Behavior**: Fully Asynchronous (everything buffered).
*   **Sinks**: Console (Async), File (Async), Network (Async).
*   **Reliability**: Non-blocking, best-effort delivery.

### 4. Developer (`NewDevelLogger`)
*   **Target**: Local Development.
*   **Behavior**: Synchronous.
*   **Sinks**: Console (Text), File (Text/Readable).

### 5. Minimal (`NewMinimalLogger`)
*   **Target**: Lightweight Applications / CLIs.
*   **Behavior**: Async Console only.
*   **Sinks**: Console (Async).
*   **Dependencies**: No external config or detailed file logging needed.

### 6. Notification Logger (`NewNotifLogger`)
*   **Target**: Services that need custom handling of alerts (e.g., Notification Servers).
*   **Behavior**: Similar to `NoLockLogger` (Fully Async).
*   **Notifier**: Uses a **Local Notifier** (Channel) instead of sending alerts over the network.
*   **API**: Exposes `SetLocalNotifQueue(chan *models.NotifMessage)` to bind the alert stream.

## Usage

```go
package main

import (
    "github.com/Bastien-Antigravity/flexible-logger/src/profiles"
    distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

func main() {
    // 1. Load Config
    config := distributed_config.New("standalone")

    // 2. Initialize Logger
    logger := profiles.NewHighPerfLogger("MyApp", config)
    defer logger.Close()

    // 3. Log
    logger.Info("Application started")
    logger.Debug("Processing request", "requestID", 123)
    
    // 4. Notifications (Warnings/Errors trigger this automatically if configured)
    logger.Warning("High memory usage detected")
}
```
