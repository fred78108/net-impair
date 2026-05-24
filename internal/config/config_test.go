package config_test

import (
	"sync"
	"testing"

	"github.com/fred78108/net-impair/internal/config"
)

func TestNewConfigDefaults(t *testing.T) {
	cfg := config.NewConfig()
	if cfg.LossPercent != 0.0 {
		t.Errorf("LossPercent: got %v, want 0.0", cfg.LossPercent)
	}
	if cfg.LatencyMs != 0 {
		t.Errorf("LatencyMs: got %v, want 0", cfg.LatencyMs)
	}
	if cfg.JitterModel != "uniform" {
		t.Errorf("JitterModel: got %q, want %q", cfg.JitterModel, "uniform")
	}
	if cfg.JitterShape != 1.0 {
		t.Errorf("JitterShape: got %v, want 1.0", cfg.JitterShape)
	}
	if cfg.JitterMix != 0.5 {
		t.Errorf("JitterMix: got %v, want 0.5", cfg.JitterMix)
	}
}

func TestValidate(t *testing.T) {
	valid := func() config.Config { return *config.NewConfig() }

	cases := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		// baseline
		{"defaults are valid", valid(), false},

		// LossPercent boundaries
		{"loss 0", func() config.Config { c := valid(); c.LossPercent = 0; return c }(), false},
		{"loss 100", func() config.Config { c := valid(); c.LossPercent = 100; return c }(), false},
		{"loss negative", func() config.Config { c := valid(); c.LossPercent = -0.1; return c }(), true},
		{"loss above 100", func() config.Config { c := valid(); c.LossPercent = 100.1; return c }(), true},

		// LatencyMs
		{"latency zero", func() config.Config { c := valid(); c.LatencyMs = 0; return c }(), false},
		{"latency positive", func() config.Config { c := valid(); c.LatencyMs = 200; return c }(), false},
		{"latency negative", func() config.Config { c := valid(); c.LatencyMs = -1; return c }(), true},

		// JitterModel valid values
		{"model uniform", func() config.Config { c := valid(); c.JitterModel = "uniform"; return c }(), false},
		{"model normal", func() config.Config { c := valid(); c.JitterModel = "normal"; return c }(), false},
		{"model pareto", func() config.Config { c := valid(); c.JitterModel = "pareto"; return c }(), false},
		{"model pareto_normal", func() config.Config { c := valid(); c.JitterModel = "pareto_normal"; return c }(), false},
		{"model gilbert_elliott", func() config.Config { c := valid(); c.JitterModel = "gilbert_elliott"; return c }(), false},
		{"model unknown", func() config.Config { c := valid(); c.JitterModel = "bogus"; return c }(), true},

		// Uniform: JitterMax
		{"uniform jitter_max zero", func() config.Config { c := valid(); c.JitterMax = 0; return c }(), false},
		{"uniform jitter_max negative", func() config.Config { c := valid(); c.JitterMax = -1; return c }(), true},

		// Normal: JitterStddev
		{"normal stddev zero", func() config.Config {
			c := valid(); c.JitterModel = "normal"; c.JitterStddev = 0; return c
		}(), false},
		{"normal stddev negative", func() config.Config {
			c := valid(); c.JitterModel = "normal"; c.JitterStddev = -1; return c
		}(), true},

		// Pareto: JitterShape
		{"pareto shape positive", func() config.Config {
			c := valid(); c.JitterModel = "pareto"; c.JitterShape = 0.5; return c
		}(), false},
		{"pareto shape zero", func() config.Config {
			c := valid(); c.JitterModel = "pareto"; c.JitterShape = 0; return c
		}(), true},
		{"pareto shape negative", func() config.Config {
			c := valid(); c.JitterModel = "pareto"; c.JitterShape = -1; return c
		}(), true},

		// ParetoNormal: JitterMix
		{"pareto_normal mix 0", func() config.Config {
			c := valid(); c.JitterModel = "pareto_normal"; c.JitterMix = 0; return c
		}(), false},
		{"pareto_normal mix 1", func() config.Config {
			c := valid(); c.JitterModel = "pareto_normal"; c.JitterMix = 1; return c
		}(), false},
		{"pareto_normal mix negative", func() config.Config {
			c := valid(); c.JitterModel = "pareto_normal"; c.JitterMix = -0.1; return c
		}(), true},
		{"pareto_normal mix above 1", func() config.Config {
			c := valid(); c.JitterModel = "pareto_normal"; c.JitterMix = 1.1; return c
		}(), true},

		// Gilbert-Elliott: delays and transition probabilities
		{"gilbert_elliott valid", func() config.Config {
			c := valid()
			c.JitterModel = "gilbert_elliott"
			c.GoodDelay, c.BadDelay = 10, 100
			c.PGoodToBad, c.PBadToGood = 0.1, 0.9
			return c
		}(), false},
		{"gilbert_elliott good_delay negative", func() config.Config {
			c := valid(); c.JitterModel = "gilbert_elliott"; c.GoodDelay = -1; return c
		}(), true},
		{"gilbert_elliott bad_delay negative", func() config.Config {
			c := valid(); c.JitterModel = "gilbert_elliott"; c.BadDelay = -1; return c
		}(), true},
		{"gilbert_elliott p_good_to_bad above 1", func() config.Config {
			c := valid(); c.JitterModel = "gilbert_elliott"; c.PGoodToBad = 1.1; return c
		}(), true},
		{"gilbert_elliott p_bad_to_good negative", func() config.Config {
			c := valid(); c.JitterModel = "gilbert_elliott"; c.PBadToGood = -0.1; return c
		}(), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestStoreUpdateVisible(t *testing.T) {
	store := config.NewStore(*config.NewConfig())
	if err := store.Update(func(c *config.Config) {
		c.LossPercent = 42.0
		c.LatencyMs = 100
		c.JitterModel = "normal"
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := store.Load()
	if got.LossPercent != 42.0 {
		t.Errorf("LossPercent: got %v, want 42.0", got.LossPercent)
	}
	if got.LatencyMs != 100 {
		t.Errorf("LatencyMs: got %v, want 100", got.LatencyMs)
	}
	if got.JitterModel != "normal" {
		t.Errorf("JitterModel: got %q, want %q", got.JitterModel, "normal")
	}
}

func TestStoreUpdateInvalidRejected(t *testing.T) {
	store := config.NewStore(*config.NewConfig())
	err := store.Update(func(c *config.Config) {
		c.LossPercent = 999
	})
	if err == nil {
		t.Fatal("expected error for invalid LossPercent, got nil")
	}
	got := store.Load()
	if got.LossPercent != 0.0 {
		t.Errorf("store was mutated after invalid update: LossPercent = %v", got.LossPercent)
	}
}

func TestStoreNoConccurentRace(t *testing.T) {
	store := config.NewStore(*config.NewConfig())
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = store.Load()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 100; j++ {
			// valid range only so updates actually commit
			_ = store.Update(func(c *config.Config) { c.LossPercent = float64(j) })
		}
	}()
	wg.Wait()
}
