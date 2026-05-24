# net-impair

!!! warning
Not ready for use. In early stages of development.

A cross-platform network impairment tool for testing applications under degraded network conditions. It creates a virtual TUN network interface and applies configurable packet loss, latency, and jitter to all traffic routed through it. A browser-based UI lets you adjust impairments in real time while your application is running.

## Motivation

Earlier in my career I worked as a network engineer. I was fortunate to work with gear that let me introduce network impairments. This helped me test not only network configurations under adverse conditions but also applications that rely on network traversal such as VoIP.

There are tools out there but not none that hit the mark on what I was looking for. This need and a desire to code outside of Python (my primary language over the last few years) motivated this project.

## Status - Under development

This project is under development. Please see [PLAN.md](PLAN.md) for the current status. The overall design and features are subject to change as I dive into this project.

## Features

- **Packet loss** — drop a configurable percentage of packets
- **Latency** — add a fixed base delay to every packet
- **Jitter** — add variable delay using one of five distribution models (Uniform, Normal, Pareto, Pareto-Normal, Gilbert-Elliott)
- **Live configuration** — change any parameter instantly from the web UI; no restart required
- **Raw IP level** — impairments apply to actual IP packets, so TCP retransmits and congestion control behave realistically
- **Cross-platform** — Linux, macOS, Windows (single binary per platform)

## How It Works

net-impair creates a TUN virtual network interface. You then add OS routing rules to direct the traffic you want to impair through that interface. Packets enter the TUN interface, pass through the impairment engine, and are forwarded to the real network. The web UI (default port 8080) controls the impairment parameters.

```
Application → OS routing table → TUN interface → Impairment engine → Physical NIC → Network
```

## Prerequisites

| Platform | Requirement                                              |
| -------- | -------------------------------------------------------- |
| Linux    | Root or `CAP_NET_ADMIN` capability                       |
| macOS    | Root (`sudo`)                                            |
| Windows  | Administrator; `wintun.dll` (bundled in the release zip) |

## Building

```sh
# Install Go 1.22+ and Node.js 20+
make build        # builds frontend then compiles Go binary
make build-linux  # cross-compile for Linux
make build-mac    # cross-compile for macOS
make build-win    # cross-compile for Windows
```

## Running

```sh
# Linux / macOS
sudo ./net-impair

# Windows (Administrator terminal)
net-impair.exe
```

The web UI is available at `http://localhost:8080` once the process starts. The TUN interface name is printed to stdout on startup (e.g., `net-impair0` on Linux, `utun8` on macOS).

---

## Routing Traffic Through the Virtual NIC

After starting net-impair, add OS routing rules to send your target traffic through the TUN interface. The tool assigns itself the address `10.0.0.1/30`; the gateway address for routing purposes is `10.0.0.2`.

> **Tip**: route only the specific IP or CIDR you are testing rather than all traffic. Routing all traffic through the TUN while also needing to reach the internet requires split-routing to avoid loops (your physical gateway must remain reachable via the physical NIC).

---

### Linux

```sh
# Replace eth0 with your physical interface and 203.0.113.5 with your target IP or CIDR

# Bring up the TUN interface (done automatically by net-impair, shown for reference)
# sudo ip addr add 10.0.0.1/30 dev net-impair0
# sudo ip link set net-impair0 up

# Route a specific target through the TUN
sudo ip route add 203.0.113.5/32 dev net-impair0

# Route a whole subnet through the TUN
sudo ip route add 203.0.113.0/24 dev net-impair0

# Route ALL traffic through the TUN (split-routing: keep physical gateway reachable)
GATEWAY=$(ip route show default | awk '/default/ {print $3}')
sudo ip route add "$GATEWAY" via "$GATEWAY" dev eth0   # anchor physical gateway
sudo ip route add default dev net-impair0

# Remove a route when done
sudo ip route del 203.0.113.5/32
```

Enable IP forwarding if you see packets not passing through:

```sh
sudo sysctl -w net.ipv4.ip_forward=1
```

---

### macOS

```sh
# The TUN interface name is printed on startup, e.g. utun8
# Replace utun8 with the name shown and 203.0.113.5 with your target

# Route a specific target through the TUN
sudo route add -host 203.0.113.5 -interface utun8

# Route a whole subnet
sudo route add -net 203.0.113.0/24 -interface utun8

# Route ALL traffic (split-routing: keep physical gateway reachable first)
GATEWAY=$(route -n get default | awk '/gateway/ {print $2}')
IFACE=$(route -n get default | awk '/interface/ {print $2}')
sudo route add "$GATEWAY" -interface "$IFACE"     # anchor physical gateway
sudo route add default -interface utun8

# Remove a route when done
sudo route delete -host 203.0.113.5
```

---

### Windows

Run all commands in an **Administrator** command prompt or PowerShell.

```cmd
REM Find the TUN interface index printed on startup, or look in Device Manager
REM Replace 203.0.113.5 and the mask with your target

REM Route a specific host through the TUN (metric 1 ensures preference)
route add 203.0.113.5 mask 255.255.255.255 10.0.0.2 metric 1

REM Route a subnet
route add 203.0.113.0 mask 255.255.255.0 10.0.0.2 metric 1

REM Route ALL traffic (split-routing: anchor physical gateway first)
REM Find your current default gateway:
ipconfig
REM Then:
route add <your-gateway-ip> mask 255.255.255.255 <your-gateway-ip>
route add 0.0.0.0 mask 0.0.0.0 10.0.0.2 metric 1

REM Remove a route when done
route delete 203.0.113.5
```

PowerShell equivalent:

```powershell
# Add route
New-NetRoute -DestinationPrefix "203.0.113.5/32" -InterfaceAlias "net-impair" -NextHop "10.0.0.2" -RouteMetric 1

# Remove route
Remove-NetRoute -DestinationPrefix "203.0.113.5/32" -Confirm:$false
```

---

## Jitter Models

Select a model in the web UI. Each model exposes only the parameters it uses.

| Model               | Description                                       | Good for                   |
| ------------------- | ------------------------------------------------- | -------------------------- |
| **Uniform**         | Random delay in `[0, max]`                        | Simple testing             |
| **Normal**          | Gaussian delay around a mean                      | Stable Wi-Fi, wired LAN    |
| **Pareto**          | Heavy-tailed; occasional large spikes             | DSL, cable                 |
| **Pareto-Normal**   | Mix of Normal and Pareto (tc netem default)       | General internet path      |
| **Gilbert-Elliott** | Two-state Markov (good/bad state machine); bursty | Mobile (LTE/5G), satellite |

See [ADR-005](docs/architecture/decisions/ADR-005-jitter-distribution-models.md) for full parameter descriptions.

---

## Architecture

Design decisions are recorded in [docs/architecture/decisions/](docs/architecture/decisions/).
System diagrams are in [docs/architecture/diagrams/](docs/architecture/diagrams/).

---

## Third-Party Licenses

This software bundles [Wintun](https://www.wintun.net) on Windows builds.
Wintun is copyright © 2018–2021 WireGuard LLC and is used under the MIT License.
See [third_party/wintun/LICENSE](third_party/wintun/LICENSE).

---

## License

MIT — see [LICENSE](LICENSE).
