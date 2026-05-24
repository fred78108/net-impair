# Architecture Diagrams

All diagrams are written in [Mermaid](https://mermaid.js.org/) and embedded directly in this file.

> **Viewing:** VS Code renders these with the [Markdown Preview Mermaid Support](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid) extension. GitHub renders them natively.

## System Overview

High-level component view of the running system.

```mermaid
flowchart TD
    subgraph host["Host Machine"]
        app["Application Under Test"]
        os_route["OS Routing Table"]
        tun["TUN Virtual NIC\n(net-impair)"]
        impair["Impairment Engine\n(latency · jitter · loss)"]
        web["Web UI\n(React — port 8080)"]
        real_nic["Physical NIC"]
    end

    network["External Network / Target Services"]

    app -->|"outbound traffic"| os_route
    os_route -->|"routed through TUN"| tun
    tun --> impair
    impair -->|"impaired packets"| real_nic
    real_nic <-->|"TCP/IP"| network

    web -->|"configure impairments via REST API"| impair

    style tun fill:#d0e8ff,stroke:#3a8fd1
    style impair fill:#ffe8cc,stroke:#d18a3a
    style web fill:#e8ffd0,stroke:#5aa13a
```
