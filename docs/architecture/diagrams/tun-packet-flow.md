# TUN Packet Flow

End-to-end path of an outbound IP packet from the application through the impairment engine to the network. Return (inbound) traffic follows normal OS routing and bypasses the TUN interface.

Related: [ADR-002](../decisions/ADR-002-proxy-approach-tun-virtual-nic.md)

```mermaid
flowchart TD
    app["Application Under Test"]
    routing["OS Routing Table\n(user-configured rules)"]
    tun["TUN Virtual NIC\nnet-impair0 · utun8 · net-impair"]
    engine["Impairment Engine"]
    loss{"Drop?\nrand() ≤ loss_pct"}
    discard(["Discard"])
    delay["Sleep\nbase_latency + jitter_sample"]
    nic["Physical NIC"]
    network["External Network\nTarget Services"]

    app -->|"outbound IP packets"| routing
    routing -->|"matched route → TUN"| tun
    tun -->|"raw IP packet"| engine
    engine --> loss
    loss -->|"yes"| discard
    loss -->|"no"| delay
    delay -->|"impaired packet"| nic
    nic -->|"send"| network

    network -->|"inbound packets"| nic
    nic -->|"normal OS routing\n(bypasses TUN)"| routing
    routing -->|"deliver"| app

    style tun fill:#d0e8ff,stroke:#3a8fd1
    style engine fill:#ffe8cc,stroke:#d18a3a
    style discard fill:#ffd0d0,stroke:#d13a3a
```
