# ADR-002: Proxy Approach — TUN Virtual NIC

## Status
Accepted

## Context
Two viable userspace approaches exist for intercepting and impairing traffic:

1. **SOCKS5 proxy** — apps are configured to route through the proxy; impairments applied to stream data; no kernel driver required
2. **TUN virtual NIC** — a virtual network interface is created; OS routing table directs traffic through it; impairments applied at the IP packet level

The primary use case is testing application behavior under realistic network conditions.

## Decision
Use a TUN virtual NIC.

## Rationale
- Packet-level impairment is more realistic: dropped packets trigger genuine TCP retransmits and congestion control responses, which is what we want to observe
- Transparent to the application under test — no proxy configuration needed in the app
- Consistent with how tools like `tc netem` (Linux) and WireGuard work conceptually
- SOCKS5 loss simulation operates on stream chunks, not IP packets, producing results that don't reflect real network behavior

## Consequences
- Requires elevated privileges (root on Linux/macOS, Administrator on Windows)
- Windows requires the Wintun kernel driver (`wintun.dll`) — MIT licensed, can be bundled; this is a prerequisite that must be documented and distributed
- More complex packet processing pipeline than a stream proxy
- A userspace networking stack decision is required to process packets above IP (see open questions)
- Routing configuration (directing OS traffic through the TUN interface) must be addressed — either automated or documented as a manual step

## Resolved Questions

**Packet processing level**: Operate at raw IP packet level — no userspace TCP/IP stack. Each IP packet is treated as an opaque unit for the purposes of delay, drop, and jitter decisions. This preserves true TCP behavior (retransmits, congestion control) because the OS stack on the far side handles reassembly, not this tool. See ADR-005 for the impairment models applied at this level.

**Routing configuration**: Manual, with documented per-OS instructions. The tool creates and configures the TUN interface; the user adds OS routing rules to direct target traffic through it. This avoids the tool needing to modify routing tables (which would require broader privileges and carries risk of leaving the host in a broken state on crash). Clear instructions are provided in the README for Linux, macOS, and Windows.
