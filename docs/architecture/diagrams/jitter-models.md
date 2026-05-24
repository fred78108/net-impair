# Jitter Distribution Models

The active jitter model is selected in the web UI. Each model is independently testable and exposes only its relevant parameters.

Related: [ADR-005](../decisions/ADR-005-jitter-distribution-models.md)

## Model Selection

```mermaid
flowchart LR
    select["Active Model\n(config.jitter_model)"]

    subgraph u["Uniform"]
        ud["rand(0, max)\n─\nParams: max (ms)\nUse: simple testing"]
    end

    subgraph n["Normal (Gaussian)"]
        nd["N(mean, σ) clamped ≥ 0\n─\nParams: mean, stddev (ms)\nUse: stable Wi-Fi · wired LAN"]
    end

    subgraph p["Pareto"]
        pd["Pareto(min, α)\nheavy tail\n─\nParams: min (ms), shape\nUse: DSL · cable"]
    end

    subgraph pn["Pareto-Normal"]
        pnd["mix×Pareto + (1-mix)×Normal\ntc netem compatible\n─\nParams: mean, stddev, shape, mix\nUse: general internet path"]
    end

    subgraph ge["Gilbert-Elliott"]
        ged["Two-state Markov\n─\nParams: good_delay, bad_delay\np_good_to_bad, p_bad_to_good\nUse: mobile (LTE/5G) · satellite"]
    end

    select -->|"UNIFORM"| u
    select -->|"NORMAL"| n
    select -->|"PARETO"| p
    select -->|"PARETO_NORMAL"| pn
    select -->|"GILBERT_ELLIOTT"| ge
```

## Gilbert-Elliott State Machine

The Gilbert-Elliott model is the only stateful model — it maintains a current state across packets to produce correlated, bursty behavior.

```mermaid
stateDiagram-v2
    direction LR
    [*] --> Good

    Good --> Good : 1 − p_good_to_bad
    Good --> Bad  : p_good_to_bad

    Bad --> Bad   : 1 − p_bad_to_good
    Bad --> Good  : p_bad_to_good

    note right of Good : delay ~ good_delay
    note right of Bad  : delay ~ bad_delay
```
