# ADR-005: Jitter Distribution Models

## Status
Accepted

## Context
Jitter is the variation in packet delay. Different network environments produce fundamentally different jitter characteristics — a corporate LAN looks nothing like a mobile LTE link or a satellite connection. A single uniform random model does not give users enough expressiveness to simulate these different environments.

## Decision
Support multiple jitter distribution models, selectable per configuration. Each model exposes only the parameters relevant to it. The web UI presents model-specific controls when a model is selected.

## Models

### 1. Uniform
The simplest model. Adds a random delay drawn uniformly from `[0, max]`.

| Parameter | Description |
|-----------|-------------|
| `max` (ms) | Upper bound of added delay |

**Suitable for**: quick sanity tests, simple impairment scenarios.

---

### 2. Normal (Gaussian)
Adds delay drawn from a normal distribution. Produces realistic jitter for stable wired or Wi-Fi connections where variation clusters around a mean.

| Parameter | Description |
|-----------|-------------|
| `mean` (ms) | Center of the distribution |
| `stddev` (ms) | Standard deviation; controls spread |

Clamped to `[0, mean + 4σ]` to avoid negative delays.

**Suitable for**: wired LAN, stable Wi-Fi.

---

### 3. Pareto
Heavy-tailed distribution that models occasional large delay spikes. More realistic for consumer broadband (DSL, cable) where the tail is long.

| Parameter | Description |
|-----------|-------------|
| `min` (ms) | Minimum delay (scale parameter) |
| `shape` | Shape parameter α; lower values produce heavier tails (typical range 1.5–3.0) |

**Suitable for**: DSL, cable, asymmetric links.

---

### 4. Pareto-Normal (Mixed)
A mixture of a Pareto and a Normal distribution. This is the model used by Linux `tc netem`. It produces both a typical clustered delay and occasional heavy-tail spikes in the same session.

| Parameter | Description |
|-----------|-------------|
| `mean` (ms) | Normal component mean |
| `stddev` (ms) | Normal component standard deviation |
| `shape` | Pareto shape parameter for the spike component |
| `mix` (0–1) | Probability that a given packet uses the Pareto component |

**Suitable for**: realistic internet path simulation; recommended default for most testing.

---

### 5. Gilbert-Elliott (Two-State Markov)
A state-machine model with a "good" state and a "bad" state. The system transitions between states according to configurable probabilities. Produces correlated, bursty jitter and can model bursty packet loss simultaneously.

| Parameter | Description |
|-----------|-------------|
| `good_delay` (ms) | Mean delay in good state |
| `bad_delay` (ms) | Mean delay in bad state |
| `p_good_to_bad` (0–1) | Transition probability from good → bad state per packet |
| `p_bad_to_good` (0–1) | Transition probability from bad → good state per packet |

**Suitable for**: mobile (LTE/5G), satellite, lossy links; scenarios where bad periods cluster together.

## Rationale
- Each model adds meaningfully different behavior — they are not redundant variations of the same idea
- Named distributions (Normal, Pareto) map to known real-world environments, making it easy for users to choose
- Gilbert-Elliott is the standard academic model for bursty network behavior and is the basis for how `tc netem` models correlation
- Pareto-Normal is included because it is the `tc netem` default, giving users a familiar baseline if they have prior `netem` experience

## Consequences
- Web UI must dynamically show/hide parameter controls based on selected model
- Each model is a small, independently testable component in the impairment engine
- Future models (e.g., trace-driven replay from a pcap) can be added without changing the model-selection interface

## Open Questions
- Whether to include preset configurations (e.g., "LTE Good", "Satellite", "3G") that pre-populate model and parameters — likely a v2 concern
