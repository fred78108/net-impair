// Package config provides the shared impairment configuration and a
// concurrency-safe store for live updates.
//
// The central type is [Store], which wraps a [Config] value behind a
// read-write mutex. The packet engine calls [Store.Load] once per packet to
// obtain a snapshot; the HTTP API calls [Store.Update] when the user changes
// settings. Neither operation blocks the other for longer than a single
// struct copy.
package config

import (
	"fmt"
	"sync"
)

// Config holds all impairment parameters applied by the packet engine.
// Fields are grouped by the jitter model that uses them; only the fields
// relevant to the active [Config.JitterModel] are consulted at runtime.
// All fields carry JSON tags for direct marshal/unmarshal by the REST API.
type Config struct {
	LossPercent float64 `json:"loss_percent"` // packet drop probability, 0–100 % as a percentage, ex. 10 for 10% loss
	LatencyMs   int     `json:"latency_ms"`   // base one-way delay added to every packet, ≥ 0
	JitterModel string  `json:"jitter_model"` // one of: uniform, normal, pareto, pareto_normal, gilbert_elliott

	// Uniform model: delay sampled from [0, JitterMax] ms.
	JitterMax int `json:"jitter_max"`

	// Normal model: delay sampled from a Gaussian clamped to [0, JitterMean+4σ].
	JitterMean   int     `json:"jitter_mean"`
	JitterStddev float64 `json:"jitter_stddev"`

	// Pareto model: heavy-tailed delay; JitterShape (α) must be > 0.
	JitterShape float64 `json:"jitter_shape"`

	// ParetoNormal model: Pareto sample with probability JitterMix, Normal otherwise.
	// JitterMix must be in [0, 1].
	JitterMix float64 `json:"jitter_mix"`

	// GilbertElliott model: two-state Markov chain alternating between a
	// low-delay "good" state and a high-delay "bad" state.
	GoodDelay  int     `json:"good_delay"`    // delay applied in the good state, ms
	BadDelay   int     `json:"bad_delay"`     // delay applied in the bad state, ms
	PGoodToBad float64 `json:"p_good_to_bad"` // per-packet probability of good→bad transition
	PBadToGood float64 `json:"p_bad_to_good"` // per-packet probability of bad→good transition
}

// NewConfig returns a Config with zero impairment: no loss, no latency, and
// the uniform jitter model with sensible parameter defaults.
func NewConfig() *Config {
	return &Config{
		LossPercent:  0.0,
		LatencyMs:    0,
		JitterModel:  "uniform",
		JitterMax:    0,
		JitterMean:   0,
		JitterStddev: 0,
		JitterShape:  1.0,
		JitterMix:    0.5,
		GoodDelay:    0,
		BadDelay:     0,
		PGoodToBad:   0.0,
		PBadToGood:   0.0,
	}
}

// Store is a concurrency-safe holder for a [Config] value.
// It is safe to call [Store.Load] and [Store.Update] from multiple goroutines
// simultaneously without additional synchronisation.
type Store struct {
	mu  sync.RWMutex
	cfg Config
}

// NewStore returns a Store initialised with initial as the starting Config.
func NewStore(initial Config) *Store {
	return &Store{cfg: initial}
}

// Load returns a snapshot of the current Config.
// The caller receives a value copy and may read it freely without holding any lock.
func (s *Store) Load() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Update applies fn to a copy of the current Config, validates the result,
// and — if valid — commits it atomically. If validation fails the stored
// Config is left unchanged and the error from [Config.Validate] is returned.
// The *Config passed to fn must not be retained beyond the call.
func (s *Store) Update(fn func(*Config)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	proposed := s.cfg // copy
	fn(&proposed)
	if err := proposed.Validate(); err != nil {
		return err // original s.cfg untouched
	}
	s.cfg = proposed
	return nil
}

// Validate reports whether all fields in c are within their legal ranges.
// It returns a descriptive error for the first violation found, or nil if the
// Config is valid. Validate is called automatically by [Store.Update].
func (c *Config) Validate() error {
	if c.LossPercent < 0.0 || c.LossPercent > 100.0 {
		return fmt.Errorf("LossPercent must be between 0 and 100")
	}
	if c.LatencyMs < 0 {
		return fmt.Errorf("LatencyMs must be non-negative")
	}
	if c.JitterModel != "uniform" && c.JitterModel != "normal" && c.JitterModel != "pareto" && c.JitterModel != "pareto_normal" && c.JitterModel != "gilbert_elliott" {
		return fmt.Errorf("JitterModel must be one of: uniform, normal, pareto, pareto_normal, gilbert_elliott")
	}
	if c.JitterModel == "uniform" && c.JitterMax < 0 {
		return fmt.Errorf("JitterMax must be non-negative for uniform model")
	}
	if c.JitterModel == "normal" && (c.JitterStddev < 0) {
		return fmt.Errorf("JitterStddev must be non-negative for normal model")
	}
	if c.JitterModel == "pareto" && c.JitterShape <= 0 {
		return fmt.Errorf("JitterShape must be positive for pareto model")
	}
	if c.JitterModel == "pareto_normal" && (c.JitterMix < 0.0 || c.JitterMix > 1.0) {
		return fmt.Errorf("JitterMix must be between 0 and 1 for pareto_normal model")
	}
	if c.JitterModel == "gilbert_elliott" {
		if c.GoodDelay < 0 || c.BadDelay < 0 {
			return fmt.Errorf("GoodDelay and BadDelay must be non-negative for gilbert_elliott model")
		}
		if c.PGoodToBad < 0.0 || c.PGoodToBad > 1.0 || c.PBadToGood < 0.0 || c.PBadToGood > 1.0 {
			return fmt.Errorf("PGoodToBad and PBadToGood must be between 0 and 1 for gilbert_elliott model")
		}
	}
	return nil
}
