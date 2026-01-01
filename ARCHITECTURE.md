# Architecture

This document describes the internal design of the Flexible Logger.

## Data Flow

```mermaid
flowchart TD
    %% Styles
    classDef core fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#0d47a1;
    classDef sink fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20;
    classDef net fill:#fff8e1,stroke:#fbc02d,stroke-width:2px,color:#f57f17;
    classDef alert fill:#ffebee,stroke:#c62828,stroke-width:2px,color:#b71c1c;
    classDef pool fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px,stroke-dasharray: 5 5,color:#4a148c;

    %% Nodes
    App[Application] -->|LogEntry| Engine(LogEngine):::core
    
    subgraph Core Processing
        direction TB
        Engine -->|Filter| Decision{Level Check}:::core
        Decision -- "Rejected" --> Drop((Drop/Pool)):::pool
        Decision -- "Accepted" --> Sinks[Sink Interface]:::core
        
        %% Pool Cycle
        Drop -.-> Pool[(Sync.Pool)]:::pool
        Pool -.-> Engine
    end
    
    subgraph Sinks Pipeline
        direction TB
        Sinks --> Multi[MultiSink]:::sink
        Multi --> Console[ConsoleSink]:::sink
        Multi --> File["WriterSink (File)"]:::sink
        Multi --> Async[AsyncSink]:::sink
    end
    
    subgraph Async Network
        direction TB
        Async -- Channel --> Worker([Worker Goroutine]):::net
        Worker --> NetSink["WriterSink (Network)"]:::net
        NetSink --> Serializer["Cap'n Proto Serializer"]:::net
        Serializer --> ManagedConn[ManagedConnection]:::net
        ManagedConn --> Socket["SafeSocket / TCP"]:::net
    end
    
    subgraph Notifications
        direction TB
        Engine -. "Warning/Error" .-> Notifier[RemoteNotifier]:::alert
        Notifier -- Channel --> NotifWorker([Notifier Worker]):::alert
        NotifWorker --> NotifConn["ManagedConnection (Hello)"]:::alert
    end
```

## Key Components

### 1. LogEngine (Core)
The central entry point. It handles:
*   **Pooling**: Retrieves `LogEntry` objects from a `sync.Pool`.
*   **Filtering**: Checks log levels (Debug, Info, etc.) before processing.
*   **Routing**: Passes valid entries to the configured `Sink` and `Notifier`.

### 2. Sinks (`src/sink`)
Sinks form a pipeline to handle log data.
*   **`WriterSink`**: Wraps an `io.Writer`. Serializes the entry (e.g., to Cap'n Proto) and writes bytes.
*   **`AsyncSink`**: Decouples the application from IO. Uses a buffered channel. If the buffer is full, it drops logs to prevent blocking.
*   **`MultiSink`**: Fan-out pattern. Sends one log entry to multiple destinations (e.g., File + Network).
*   **Memory Management**: Sinks accept ownership of a `LogEntry` and are responsible for calling `Release()` to return it to the pool.

### 3. Network Manager (`src/network_manager`)
Handles robust network communication.
*   **`NetworkManager`**: Factory for creating connections.
*   **`ManagedConnection`**: A wrapper around the raw socket. It intercepts `Write()` calls; if the underlying connection is broken, it automatically attempts to reconnect (blocking or async depending on config) and retries the write.
*   **`EstablishConnection`**: Centralized logic for IP/Port resolution and socket creation.

### 4. Remote Notifier (`src/notifier`)
A separate subsystem for high-priority alerts (Warnings/Errors).
*   **Independent Channel**: Does not block the main log stream.
*   **Protocol**: Uses a dedicated lightweight protocol (`tcp-hello` profile) to send alerts to a monitoring server.
*   **Resilience**: Uses `ManagedConnection` for auto-reconnection.
