# ADR-003: Web UI — React with Vite

## Status
Accepted

## Context
The tool needs a web interface for configuring impairment parameters (latency, jitter, packet loss) and observing status. The UI must be served by the Go binary with no external dependencies at runtime.

## Decision
Build the frontend with React and Vite; embed the compiled output in the Go binary using `//go:embed`.

## Rationale
- React provides a component model suitable for controls (sliders, toggles, live stats)
- Vite produces a compact, optimized static build suitable for embedding
- The Go binary serves the embedded assets via `net/http` — no separate web server process
- A build step (npm) is required at development time but not at runtime

## Consequences
- Node.js and npm required in the development environment and CI pipeline
- Binary size increases by the size of the compiled frontend (typically 200–500 KB gzipped)
- A build script or Makefile step must compile the frontend before `go build`
- Hot-reload development workflow requires running the Vite dev server separately from the Go binary

## Open Questions
- Specific UI controls and live data to display (e.g., live throughput graph, per-connection stats)
- Whether configuration changes apply immediately to in-flight connections or only new ones
