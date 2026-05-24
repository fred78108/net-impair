# ADR-004: Impairment Granularity — Global (v1)

## Status
Accepted

## Context
Impairments (latency, jitter, packet loss) could be applied globally to all traffic passing through the TUN interface, or per-rule (e.g., per destination host, port, or protocol).

## Decision
In v1, apply a single global set of impairment parameters to all traffic through the interface.

## Rationale
- Covers the primary use case: testing one application under degraded network conditions
- Significantly simpler configuration model and implementation
- Per-rule granularity can be added in a future version without breaking the global model (global becomes the default rule)

## Consequences
- All traffic routed through the TUN interface is impaired equally — the user controls what traffic is routed through it via OS routing rules
- Per-destination or per-protocol impairment rules are deferred to a future version

## Future Consideration
Per-host or per-port rules would allow impairment of specific destinations (e.g., only impair `api.example.com`) while leaving other traffic unaffected. This is a natural v2 extension.
