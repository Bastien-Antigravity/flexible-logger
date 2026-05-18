---
microservice: flexible-logger
type: architecture
status: active
tags:
- '#ai/ignore'
- '#service/flexible-logger'
- '#domain/observability'
- '#zone/3-fleet'
---

# Architecture Overview

This document describes the high-level design, data flows, and key internal components of the Flexible Logger.

## Data Flow

```mermaid
flowchart TD
    %% Styles
    classDef core fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#0d47a1;
    classDef sink fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20;
    classDef net fill:#fff8e1,stroke:#fbc02d,stroke-width:2px,color:#f57f17;
    classDef alert fill:#ffebee,stroke:#c62828,stroke-width:2px,color:#b71c1c;
    classDef pool fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px,stroke-dasharray: 5 5,color:#4a148c;

    %% Application Node Styling
    style App fill:#37474f,stroke:#263238,stroke-width:3px,color:#ffffff

    %% Nodes
    App[Application] -->|LogEntry| Engine(LogEngine):::core
    
    subgraph Core [Core Processing]
        direction TB
        Engine -->|Filter| Decision{Level Check}:::core
        Decision -- "Rejected" --> Drop((Drop/Pool)):::pool
        Decision -- "Accepted" --> Sinks[Sink Interface]:::core
        
        %% Pool Cycle
        Drop -.-> Pool[(Sync.Pool)]:::pool
        Pool -.-> Engine
    end
    style Core fill:#edf7ff,stroke:#82b1ff,stroke-width:2px,color:#0d47a1
    
    subgraph SinksPipe [Sinks Pipeline]
        direction TB
        Sinks --> Multi[MultiSink]:::sink
        Multi --> Console[ConsoleSink]:::sink
        Multi --> File["WriterSink (File)"]:::sink
        Multi --> Async[AsyncSink]:::sink
    end
    style SinksPipe fill:#f1f8e9,stroke:#aed581,stroke-width:2px,color:#33691e
    
    subgraph AsyncNet [Async Network]
        direction TB
        Async -- Channel --> Worker([Worker Goroutine]):::net
        Worker --> NetSink["WriterSink (Network)"]:::net
        NetSink --> Serializer["Cap'n Proto Serializer"]:::net
        Serializer --> ManagedConn[ManagedConnection]:::net
        ManagedConn --> Socket["SafeSocket / TCP"]:::net
    end
    style AsyncNet fill:#fffde7,stroke:#fff176,stroke-width:2px,color:#f57f17
    
    subgraph Notif [Notifications]
        direction TB
        Engine -. "Warning/Error" .-> Notifier[Notifier Interface]:::alert
        Notifier -- "Remote" --> Remote[RemoteNotifier]:::alert
        Notifier -- "Local" --> Local[LocalNotifier]:::alert
        Remote -- Channel --> NotifWorker([Notifier Worker]):::alert
        NotifWorker --> NotifConn["ManagedConnection (Hello)"]:::alert
        Local -- "Go Channel" --> App
    end
    style Notif fill:#ffebee,stroke:#ffcdd2,stroke-width:2px,color:#b71c1c
```

## Key Components

### 1. LogEngine (Core)
The central entry point for all logging calls:
*   **Printf-style API**: Methods accept `(format string, args ...any)`, formatting the message internally via Go's standard formatting library.
*   **Pooling**: Leverages a robust `sync.Pool` to reuse `LogEntry` structures, avoiding unnecessary allocations on hot paths.
*   **Filtering**: Checks active log levels (e.g., Debug, Info, etc.) before proceeding with deeper formatting or caller discovery.
*   **Sampling**: Employs probabilistic dropping strategies on high-throughput non-critical paths to protect CPU/memory resources without losing important state.
*   **Metadata Enrichment**: Automatically tags logs with standard context including `ProcessID`, `ProcessName`, and `Hostname`.
*   **Smart Caller Discovery**: Uses `runtime.Caller` selectively based on log level or profile settings to locate code source file/line numbers efficiently.
*   **Routing**: Handles internal dispatching of entries to the correct `Sink` pipeline and optional `Notifier` channels.

### 2. Sinks Pipeline
All logs flow from the `LogEngine` through standard sink pipelines:
*   **`WriterSink`**: Wraps standard writers (like Console or Files). Serializes logs (e.g., into Cap'n Proto bytes) and executes the write.
*   **`AsyncSink`**: Decouples active execution threads from blocking I/O using highly-optimized buffered channels. Supports smart dropping behaviors if capacity thresholds are exceeded under high pressure.
*   **`MultiSink`**: Supports fan-out operations, distributing log payloads synchronously/asynchronously across multiple underlying sinks (e.g., local console logging combined with remote network logging).

### 3. Connection Management (Outsourced)
In alignment with modern microservice standards, the logger delegates underlying TCP socket and lifecycle management:
*   **Centralized Integration**: Leverages the `conn_manager.NetworkManager` from the `microservice-toolbox`.
*   **ManagedConnection**: Automatically handles standard backoff profiles, retry jittering, connection state bookkeeping, and standard handshakes.

### 4. Notifiers Subsystem
A specialized pathway for high-importance telemetry (such as Warnings or Critical Errors):
*   **Decoupled Path**: Avoids the standard sink queues to guarantee low latency.
*   **`RemoteNotifier`**: Employs a dedicated TCP profile (`tcp-hello`) for sending alerts to target monitor processes, serializing payloads using standard Cap'n Proto schemas.
*   **`LocalNotifier`**: Pipes alerts to Go channels in-process, allowing local applications to react dynamically to error states.
