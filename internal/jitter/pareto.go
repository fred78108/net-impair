package jitter

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// paretoSampler produces heavy-tailed delays using the Pareto distribution.
// The scale/minimum parameter maps to cfg.JitterMean; see ADR-007 for the full
// config field mapping rationale.
// Registered under the model name "pareto".
type paretoSampler struct {
	min   float64 // ms — sourced from cfg.JitterMean
	shape float64 // α — sourced from cfg.JitterShape
}

// newPareto constructs a paretoSampler from cfg.JitterMean (min) and cfg.JitterShape (α).
func newPareto(cfg config.Config) (Sampler, error) {
	return &paretoSampler{
		min:   float64(cfg.JitterMean),
		shape: cfg.JitterShape,
	}, nil
}

// Sample returns a heavy-tailed delay using the inverse-CDF method: min / U^(1/α),
// where U is uniform on (0, 1].
func (s *paretoSampler) Sample() time.Duration {
	u := rand.Float64()
	v := s.min / math.Pow(u, 1.0/s.shape)
	return time.Duration(v * float64(time.Millisecond))
}
