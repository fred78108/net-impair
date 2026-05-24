// Package jitter implements per-packet delay sampling for the packet engine.
// Each jitter model is selected by name at config-change time via [New];
// the returned [Sampler] is then called once per packet on the hot path.
package jitter

import (
	"fmt"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// Sampler is the interface implemented by all jitter models.
// It is called once per packet on the hot path; implementations must be safe
// for concurrent use.
type Sampler interface {
	// Sample returns the jitter delay to add to a single packet.
	Sample() time.Duration
}

// Constructor is the signature of a model constructor stored in the registry.
// It receives the full config and extracts whichever fields are relevant to
// that model.
type Constructor func(cfg config.Config) (Sampler, error)

var registry = map[string]Constructor{
	"uniform":         newUniform,
	"normal":          newNormal,
	"pareto":          newPareto,
	"pareto_normal":   newParetoNormal,
	"gilbert_elliott": newGilbertElliott,
}

// New returns a [Sampler] for the model named in cfg.JitterModel.
// Valid model names are: uniform, normal, pareto, pareto_normal, gilbert_elliott.
// New is called once when the configuration changes, not per packet.
func New(cfg config.Config) (Sampler, error) {
	ctor, ok := registry[cfg.JitterModel]
	if !ok {
		return nil, fmt.Errorf("unknown jitter model %q", cfg.JitterModel)
	}
	return ctor(cfg)
}
