package jitter

import (
	"math/rand/v2"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// uniformSampler produces delays drawn uniformly from [0, max] milliseconds.
// Registered under the model name "uniform".
type uniformSampler struct {
	max int // ms
}

// newUniform constructs a uniformSampler from cfg.JitterMax.
func newUniform(cfg config.Config) (Sampler, error) {
	return &uniformSampler{max: cfg.JitterMax}, nil
}

// Sample returns a delay uniformly distributed in [0, max] ms.
func (s *uniformSampler) Sample() time.Duration {
	if s.max == 0 {
		return 0
	}
	return time.Duration(rand.IntN(s.max+1)) * time.Millisecond
}
