# Architecture Diagrams

| Diagram | Description |
|---------|-------------|
| [System Overview](system-overview.md) | High-level component view of the running system |
| [TUN Packet Flow](tun-packet-flow.md) | End-to-end path of an IP packet through the impairment engine |
| [Impairment Engine](impairment-engine.md) | Per-packet processing pipeline and concurrent config access pattern |
| [Jitter Models](jitter-models.md) | Five jitter distribution models and the Gilbert-Elliott state machine |
| [Live Config Update](config-live-update.md) | Sequence for changing impairment parameters via the web UI at runtime |
| [Build Pipeline](build-pipeline.md) | Frontend compilation, Go embed, and cross-platform binary targets |
| [Startup Sequence](startup-sequence.md) | Initialization steps from binary launch to first packet processed |
