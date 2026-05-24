# ADR-001: Implementation Language — Go

## Status
Accepted

## Context
The tool must run on Windows, macOS, and Linux without requiring a runtime installation. It involves low-level networking (TUN interfaces, raw packet handling) and needs to ship as a single self-contained binary. A web UI will be embedded in the binary.

## Decision
Use Go.

## Rationale
- Cross-platform compilation to a single static binary with no runtime dependency
- Strong standard library for networking and HTTP
- Native support for embedding static assets (`//go:embed`)
- The WireGuard project (closest reference implementation for TUN on all three platforms) is written in Go — relevant libraries are idiomatic Go
- Sufficient ecosystem for TUN/TAP (`golang.zx2c4.com/wireguard/tun`, `github.com/songgao/water`) and userspace networking

## Consequences
- Wintun driver dependency on Windows (see ADR-002)
- Team must be comfortable with Go; no scripting-language flexibility
