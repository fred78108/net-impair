# Live Configuration Update

Sequence for changing impairment parameters via the web UI while the proxy is running. Changes take effect immediately on the next packet processed — there is no per-connection state or drain period.

Related: [ADR-006](../decisions/ADR-006-live-configuration-changes.md)

```mermaid
sequenceDiagram
    actor Eng as Engineer
    participant UI as Browser (React UI)
    participant API as Go HTTP Handler
    participant Cfg as Config Struct (sync.RWMutex)
    participant Pipe as Packet Processing Goroutine

    Eng->>UI: Adjust slider / toggle
    UI->>API: PUT /api/config<br/>{latency_ms, loss_pct, jitter_model, ...}
    API->>Cfg: Lock() — exclusive write
    API->>Cfg: cfg = newValues
    API->>Cfg: Unlock()
    API-->>UI: 200 OK

    Note over Pipe,Cfg: Concurrent — next packet arrives
    Pipe->>Cfg: RLock() — shared read
    Cfg-->>Pipe: config snapshot
    Pipe->>Cfg: RUnlock()
    Pipe->>Pipe: Apply new latency / jitter / loss
```

> **Note**: individual packets already in OS buffers when the config write occurs may still complete with the prior timing, but no batching or drain logic is required — correctness is guaranteed at the struct level.
