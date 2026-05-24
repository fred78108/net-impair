# ADR-006: Live Configuration Changes

## Status
Accepted

## Context
Impairment parameters (latency, jitter model, packet loss) can be changed via the web UI while the proxy is running. The question is whether those changes take effect immediately on all traffic or only on connections established after the change.

## Decision
Configuration changes take effect immediately — the next packet processed after the UI applies a change will use the new parameters. There is no concept of "pending" or "per-connection" configuration.

## Rationale
- The primary use case is interactive testing: an engineer observes the application under test, adjusts impairments in real time, and immediately observes the effect
- Immediate application is the only model that supports this workflow
- At the raw IP packet level (see ADR-002), there is no persistent per-connection state in the impairment engine — each packet is processed independently against the current configuration, so immediate application is natural
- A "new connections only" model would require tracking per-flow state, adding complexity with no clear benefit

## Implementation Note
Configuration is stored in a single in-memory struct protected by a read-write mutex. The packet processing pipeline holds a read lock per packet; the web API handler holds a write lock when applying changes. This ensures consistency without pausing the pipeline for longer than a single config read.

## Consequences
- Users can observe in real time how an application responds to changing conditions (e.g., gradually increasing latency)
- No warmup or drain period when changing settings — the change is atomic at the config-struct level but individual in-flight packets in OS buffers may still use old timing
- Configuration is not persisted across restarts in v1; the tool starts with zero impairment each time

## Open Questions
- Whether to support named configuration snapshots (save/load presets) — deferred to a future version
