package jitter

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// paretoNormalSampler produces delays from a mixture of Pareto and Normal
// distributions, matching the model used by Linux tc-netem. Each call takes
// the Pareto branch with probability mix and the Normal branch otherwise.
// Registered under the model name "pareto_normal".
type paretoNormalSampler struct {
	mean   float64 // ms
	stddev float64 // ms
	shape  float64 // α
	mix    float64 // probability [0,1] of taking the Pareto branch per packet
}

// newParetoNormal constructs a paretoNormalSampler from cfg.JitterMean,
// cfg.JitterStddev, cfg.JitterShape, and cfg.JitterMix.
func newParetoNormal(cfg config.Config) (Sampler, error) {
	return &paretoNormalSampler{
		mean:   float64(cfg.JitterMean),
		stddev: cfg.JitterStddev,
		shape:  cfg.JitterShape,
		mix:    cfg.JitterMix,
	}, nil
}

// Sample returns a delay from the Pareto branch with probability mix, or from
// the clamped Normal branch otherwise — matching the tc-netem pareto-normal model.
func (s *paretoNormalSampler) Sample() time.Duration {
	var v float64
	if rand.Float64() < s.mix {
		u := rand.Float64()
		v = s.mean / math.Pow(u, 1.0/s.shape)
	} else {
		v = rand.NormFloat64()*s.stddev + s.mean
		if v < 0 {
			v = 0
		}
	}
	return time.Duration(v * float64(time.Millisecond))
}
