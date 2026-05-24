package jitter

import (
	"math/rand/v2"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// normalSampler produces delays drawn from a Gaussian distribution N(mean, stddev),
// clamped to ≥ 0 to prevent negative delays.
// Registered under the model name "normal".
type normalSampler struct {
	mean   float64 // ms
	stddev float64 // ms
}

// newNormal constructs a normalSampler from cfg.JitterMean and cfg.JitterStddev.
func newNormal(cfg config.Config) (Sampler, error) {
	return &normalSampler{
		mean:   float64(cfg.JitterMean),
		stddev: cfg.JitterStddev,
	}, nil
}

// Sample returns a delay drawn from N(mean, stddev), clamped to ≥ 0.
func (s *normalSampler) Sample() time.Duration {
	v := rand.NormFloat64()*s.stddev + s.mean
	if v < 0 {
		v = 0
	}
	return time.Duration(v * float64(time.Millisecond))
}
