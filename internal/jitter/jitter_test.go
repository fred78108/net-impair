package jitter

import (
	"testing"

	"github.com/fred78108/net-impair/internal/config"
)

func TestNew_UnknownModel(t *testing.T) {
	_, err := New(config.Config{JitterModel: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown model, got nil")
	}
}

func TestNew_AllModels(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.Config
	}{
		{"uniform", config.Config{JitterModel: "uniform", JitterMax: 10}},
		{"normal", config.Config{JitterModel: "normal", JitterMean: 20, JitterStddev: 5}},
		{"pareto", config.Config{JitterModel: "pareto", JitterMean: 5, JitterShape: 2}},
		{"pareto_normal", config.Config{JitterModel: "pareto_normal", JitterMean: 20, JitterStddev: 5, JitterShape: 2, JitterMix: 0.5}},
		{"gilbert_elliott", config.Config{JitterModel: "gilbert_elliott", GoodDelay: 5, BadDelay: 50, PGoodToBad: 0.1, PBadToGood: 0.5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(tc.cfg)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}
			if s == nil {
				t.Fatal("New() returned nil Sampler")
			}
		})
	}
}
