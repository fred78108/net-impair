# net-impair Development Plan

## Progress

- [ ] Phase 1 — Go Module Setup
- [ ] Phase 2 — `internal/config`
- [ ] Phase 3 — `internal/jitter`
- [ ] Phase 4 — `internal/tun`
- [ ] Phase 5 — `internal/engine`
- [ ] Phase 6 — `internal/api`
- [ ] Phase 7 — Frontend (React + Vite)
- [ ] Phase 8 — `cmd/net-impair/main.go`
- [ ] Phase 9 — Build System
- [ ] Phase 10 — Integration Testing

---

Cross-platform network impairment tool. Creates a TUN virtual NIC and applies latency, jitter, and packet loss. Controlled via an embedded React web UI.

All packages are currently stubs. Work roughly top-to-bottom; each layer depends on the one above it.

---

## Phase 1 — Go Module Setup

- [x] Add core Go dependencies to `go.mod` / `go.sum`
  - `golang.zx2c4.com/wireguard/tun` — cross-platform TUN interface
  - `golang.org/x/net` — IP packet utilities
- [ ] Verify `go mod tidy` passes cleanly (delayed until modules implemented)

---

## Phase 2 — `internal/config`

Single source of truth for all impairment parameters. Protected by a `sync.RWMutex` so the engine can read per-packet without blocking and the API can write atomically (see ADR-006).

- [x] Define `Config` struct
  - `LossPercent float64` (0–100)
  - `LatencyMs int`
  - `JitterModel string` (uniform | normal | pareto | pareto-normal | gilbert-elliott)
  - Model-specific fields: `JitterMax`, `JitterMean`, `JitterStddev`, `JitterShape`, `JitterMix`, `GoodDelay`, `BadDelay`, `PGoodToBad`, `PBadToGood`
- [x] `NewConfig() *Config` — returns zero-impairment defaults
- [x] `Store` wrapper with `RWMutex`: `Load() Config` and `Update(fn func(*Config))` methods
- [x] JSON tags on all fields for API serialization
- [x] Unit tests: validate mutex behavior under concurrent read/write

---

## Phase 3 — `internal/jitter`

Five independently testable models. Each implements a single-method interface so the engine can swap models without a type switch (see ADR-005).

- [x] Define `Sampler` interface: `Sample() time.Duration`
- [x] Implement `uniformSampler` — `[0, max]`
- [x] Implement `normalSampler` — Gaussian, clamped to `≥ 0`
- [x] Implement `paretoSampler` — heavy-tailed, `min` (mapped from `JitterMean`) + shape α
- [x] Implement `paretoNormalSampler` — mixed Pareto/Normal with `mix` probability
- [x] Implement `gilbertElliottSampler` — two-state Markov; `sync.Mutex` guards mutable Markov state
- [x] `New(cfg config.Config) (Sampler, error)` — registry map lookup; each model registers a `Constructor` func; see ADR-007
- [x] Unit tests for each sampler: statistical sanity checks (mean within tolerance, no negative samples)

---

## Phase 4 — `internal/tun`

Wraps the wireguard/tun library to present a simple read/write packet interface. Cross-platform by the library; Windows also needs wintun.dll at runtime.

- [ ] `Device` struct wrapping `tun.TUNDevice`
- [ ] `Open(name string, addr string) (*Device, error)` — creates TUN, assigns `10.0.0.1/30`, brings interface up
- [ ] `ReadPacket(buf []byte) (int, error)` and `WritePacket(pkt []byte) error`
- [ ] Graceful `Close()` that tears down the interface
- [ ] Platform notes: name is advisory on macOS (kernel picks `utunN`), print actual name to stdout on start
- [ ] Unit tests: mock or build-tag-guarded; test open/close lifecycle

---

## Phase 5 — `internal/engine`

The hot path. Reads packets from the TUN, applies loss/latency/jitter, then writes them back out (or drops them). Each packet is processed independently against the current config snapshot (see ADR-002, ADR-006).

- [ ] `Engine` struct holding a `*tun.Device`, `*config.Store`, and a `jitter.Sampler` cache
- [ ] `Run(ctx context.Context)` — main packet loop
  - Read packet from TUN
  - Snapshot config (`store.Load()`)
  - Packet loss: drop with probability `LossPercent/100`
  - Latency + jitter: schedule write via `time.AfterFunc` with `LatencyMs + sampler.Sample()`
  - Update `jitter.Sampler` when model or params change
- [ ] Metrics counters (atomic): `PacketsIn`, `PacketsOut`, `PacketsDropped`, `BytesIn`, `BytesOut` — exposed to the API
- [ ] Graceful shutdown when `ctx` is cancelled (drain in-flight delayed packets with a short deadline)
- [ ] Unit tests: mock TUN; verify drop rate converges to configured loss %, verify packets are delayed by ≥ configured latency

---

## Phase 6 — `internal/api`

HTTP server. Serves the embedded React UI at `/` and a JSON REST API under `/api/`.

- [ ] `Server` struct holding `*config.Store` and `*engine.Engine`
- [ ] `GET /api/config` — returns current config as JSON
- [ ] `PUT /api/config` — accepts partial or full config JSON, calls `store.Update()`
- [ ] `GET /api/stats` — returns engine metrics snapshot as JSON
- [ ] `POST /api/reset` — resets config to zero impairment
- [ ] `//go:embed ui/dist` — embed compiled React assets
- [ ] Serve embedded assets at `/`; API routes take priority
- [ ] `Run(addr string) error` wrapping `http.ListenAndServe`
- [ ] Unit tests: `httptest` for each API endpoint; verify config round-trips correctly

---

## Phase 7 — Frontend (React + Vite)

Lives in `ui/`. Built with `npm run build`; output goes to `ui/dist/` which is embedded by the API package.

- [ ] Scaffold with `npm create vite@latest ui -- --template react-ts`
- [ ] **Config panel** — controls for every impairment parameter
  - Loss %: numeric input or slider (0–100)
  - Latency: numeric input (ms)
  - Jitter model: dropdown (5 options)
  - Dynamic sub-controls: show only the fields for the selected jitter model
- [ ] **Stats bar** — live display of packets in/out/dropped; poll `GET /api/stats` every second
- [ ] **Apply / Reset buttons** — `PUT /api/config` and `POST /api/reset`
- [ ] UI state reflects server state on load (fetch `GET /api/config`)
- [ ] Minimal styling — functional, not polished; no heavy component library needed
- [ ] `vite.config.ts` proxy: forward `/api` to `http://localhost:8080` during dev

---

## Phase 8 — `cmd/net-impair/main.go`

Wires everything together.

- [ ] Parse flags: `--addr` (default `0.0.0.0:8080`), `--tun-name` (default `net-impair`)
- [ ] Privilege check: warn if not running as root/admin (don't exit; let TUN open fail with a clear error)
- [ ] Open TUN device, print actual interface name to stdout
- [ ] Create config store, engine, API server
- [ ] Start engine in a goroutine; start API server in a goroutine
- [ ] Handle `SIGINT`/`SIGTERM`: cancel context, wait for engine drain, close TUN, exit cleanly

---

## Phase 9 — Build System

- [ ] Write `Makefile`
  - `make build` — `npm run build` in `ui/`, then `go build ./cmd/net-impair`
  - `make build-linux` — GOOS=linux GOARCH=amd64
  - `make build-mac` — GOOS=darwin GOARCH=arm64 (and/or amd64)
  - `make build-win` — GOOS=windows GOARCH=amd64; copy `wintun.dll` into output dir
  - `make test` — `go test ./...`
  - `make clean` — remove `ui/dist/` and built binaries
- [ ] `.gitignore`: `ui/node_modules/`, `ui/dist/`, built binaries

---

## Phase 10 — Integration Testing

- [ ] Manual smoke test on macOS: start tool, set 100ms latency, `ping` a routed host, verify RTT increase
- [ ] Manual smoke test: set 50% loss, confirm roughly half of `ping` packets drop
- [ ] Manual smoke test: change config via UI while `ping` is running, observe immediate effect
- [ ] `testdata/` — add sample pcap or packet hex for engine unit tests if needed

---

## Deferred (v2)

- Per-destination / per-port impairment rules (ADR-004)
- Named configuration presets / save-load (ADR-006 open question)
- Live throughput graph in UI (ADR-003 open question)
- Trace-driven replay from pcap (ADR-005 open question)
- Linux CI runner for automated integration tests
