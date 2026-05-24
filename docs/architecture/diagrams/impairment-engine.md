# Impairment Engine

Per-packet processing pipeline inside the impairment engine. The config struct is read under a shared read lock so config updates from the web UI never block packet processing for longer than a single struct read.

Related: [ADR-004](../decisions/ADR-004-impairment-granularity-global-v1.md) · [ADR-005](../decisions/ADR-005-jitter-distribution-models.md) · [ADR-006](../decisions/ADR-006-live-configuration-changes.md)

```mermaid
flowchart TD
    pkt_in(["IP Packet In"])
    rlock["Acquire RLock"]
    snapshot["Read config snapshot\nloss_pct · base_latency\njitter_model · model_params"]
    runlock["Release RLock"]
    loss_chk{"rand() ≤ loss_pct?"}
    discard(["Discard"])
    jitter_sample["Sample jitter delay\nfrom active model"]
    total["total_delay =\nbase_latency + jitter_sample"]
    sleep["Sleep(total_delay)"]
    forward(["Forward to Physical NIC"])

    pkt_in --> rlock
    rlock --> snapshot
    snapshot --> runlock
    runlock --> loss_chk
    loss_chk -->|"yes"| discard
    loss_chk -->|"no"| jitter_sample
    jitter_sample --> total
    total --> sleep
    sleep --> forward

    style pkt_in fill:#d0e8ff,stroke:#3a8fd1
    style discard fill:#ffd0d0,stroke:#d13a3a
    style forward fill:#d0ffd0,stroke:#3ad13a
```

## Config Struct — Concurrent Access

```mermaid
flowchart LR
    subgraph readers["Packet goroutines (many, concurrent)"]
        p1["Packet 1\nRLock → read → RUnlock"]
        p2["Packet 2\nRLock → read → RUnlock"]
    end

    subgraph writer["Web API handler (one at a time)"]
        api["PUT /api/config\nLock → write → Unlock"]
    end

    cfg[("Config Struct\nsync.RWMutex")]

    p1 <-->|"shared read"| cfg
    p2 <-->|"shared read"| cfg
    api <-->|"exclusive write"| cfg
```
