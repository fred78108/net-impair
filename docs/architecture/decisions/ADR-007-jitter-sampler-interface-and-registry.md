# ADR-007: Jitter Sampler Interface and Registry Pattern

## Status

Accepted

## Context

ADR-005 decided which jitter distribution models to support and what parameters each exposes. It did not specify how model selection is implemented inside the Go codebase — that is, how the packet engine dispatches to the correct model at runtime, how new models are added, and what concurrency contract callers can rely on.

Two common Go approaches exist:

- **Switch-based factory**: a single function with a `switch` on the model name that constructs and returns the right implementation.
- **Interface + registry map**: a named interface, a map from model name to constructor function, and a `New()` entry point that does a map lookup.

## Decision

Use a `Sampler` interface with a registry map of constructor functions.

```
internal/jitter/
  jitter.go           ← Sampler interface, Constructor type, registry map, New()
  uniform.go
  normal.go
  pareto.go
  pareto_normal.go
  gilbert_elliott.go
```

The `Sampler` interface has a single method:

```go
type Sampler interface {
    Sample() time.Duration
}
```

The registry maps model name strings to constructor functions:

```go
type Constructor func(cfg config.Config) (Sampler, error)

var registry = map[string]Constructor{ ... }
```

`New(cfg config.Config)` is the only exported entry point. It is called once when the configuration changes, not per packet.

## Rationale

**Interface + registry over switch:**
- Adding a new model touches one file (the new model file) and one map entry. A switch requires editing the factory file.
- Each constructor (`newUniform`, `newNormal`, etc.) can be called directly in tests without going through `New()`, making each model independently testable.
- The registry map is self-documenting — the full list of supported models is visible at a glance.
- In tests, a registry entry can be replaced with a fake to control sampling output deterministically.

**Single-method interface:**
- `Sample() time.Duration` is the entire contract. Config struct fields do not leak into the interface.
- The engine holds a `Sampler` and calls `Sample()` per packet with no further knowledge of which model is active.

**`New()` called at config-change time, not per packet:**
- Map lookup and constructor overhead are irrelevant to throughput.
- The engine replaces its held `Sampler` when the config changes (see ADR-006).

## Concurrency Model

All `Sampler` implementations must be safe for concurrent use. Two approaches are used depending on whether the model carries state:

**Stateless models** (Uniform, Normal, Pareto, Pareto-Normal): no fields are written after construction. These call `math/rand/v2` package-level functions directly. Since Go 1.22, the `math/rand/v2` global source is goroutine-safe, so no additional synchronisation is needed.

**Stateful models** (Gilbert-Elliott): the two-state Markov chain must persist its current state (`inBad bool`) across successive `Sample()` calls to produce correlated delay bursts. `gilbertElliottSampler` protects this field with a `sync.Mutex`.

## Config Field Mapping

The `config.Config` struct uses generic field names shared across models. The Pareto model uses `JitterMean` as its scale/minimum parameter (`min` in ADR-005) because no dedicated `JitterMin` field exists. This mapping is intentional: adding a separate field would widen the config API surface for a single model's edge case.

| ADR-005 parameter | config.Config field |
|---|---|
| Uniform `max` | `JitterMax` |
| Normal `mean` | `JitterMean` |
| Normal `stddev` | `JitterStddev` |
| Pareto `min` | `JitterMean` |
| Pareto `shape` | `JitterShape` |
| Pareto-Normal `mean` | `JitterMean` |
| Pareto-Normal `stddev` | `JitterStddev` |
| Pareto-Normal `shape` | `JitterShape` |
| Pareto-Normal `mix` | `JitterMix` |
| Gilbert-Elliott `good_delay` | `GoodDelay` |
| Gilbert-Elliott `bad_delay` | `BadDelay` |
| Gilbert-Elliott `p_good_to_bad` | `PGoodToBad` |
| Gilbert-Elliott `p_bad_to_good` | `PBadToGood` |

## Consequences

- The engine depends only on the `Sampler` interface, not on any concrete model type.
- Each model file is independently readable and testable with no knowledge of the others.
- Future models (e.g., trace-driven replay) require: a new file, a new constructor, and one map entry — no changes to existing code.
- The `JitterMean` dual-use for Normal mean and Pareto min is a known ambiguity; it is documented here and in `jitter-models.md` to prevent confusion.
