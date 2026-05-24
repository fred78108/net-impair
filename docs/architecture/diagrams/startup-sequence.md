# Startup Sequence

Initialization steps from binary launch to ready state. After the tool prints its ready message, the operator must manually add OS routing rules to direct target traffic through the TUN interface.

Related: [ADR-002](../decisions/ADR-002-proxy-approach-tun-virtual-nic.md) · [ADR-003](../decisions/ADR-003-frontend-react-vite.md)

```mermaid
sequenceDiagram
    actor Op as Operator
    participant Bin as net-impair binary
    participant OS as Operating System
    participant TUN as TUN Interface
    participant HTTP as HTTP Server (port 8080)
    participant PktLoop as Packet Processing Goroutine

    Op->>Bin: sudo ./net-impair
    Bin->>OS: Create TUN interface<br/>(requires root / CAP_NET_ADMIN / Administrator)
    OS-->>Bin: Interface name (net-impair0, utun8, net-impair)
    Bin->>TUN: Assign 10.0.0.1/30, bring up
    Bin->>HTTP: Start listener<br/>serve embedded frontend (//go:embed)
    Bin->>PktLoop: Start goroutine — read packets from TUN fd
    Bin-->>Op: stdout: TUN interface: net-impair0<br/>Web UI: http://localhost:8080

    Note over Op,OS: Manual routing step (per-OS instructions in README)
    Op->>OS: ip route add <target> dev net-impair0
    OS-->>TUN: Traffic for <target> now flows through TUN
    Note over PktLoop,TUN: Impairment engine begins processing packets
```
