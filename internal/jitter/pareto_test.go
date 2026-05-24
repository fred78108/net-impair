package jitter

import (
	"testing"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

func TestPareto_AboveMin(t *testing.T) {
	const min = 10
	s, _ := newPareto(config.Config{JitterMean: min, JitterShape: 2})
	minDur := time.Duration(min) * time.Millisecond
	for i := 0; i < samples; i++ {
		if v := s.Sample(); v < minDur {
			t.Fatalf("sample %v below min %v", v, minDur)
		}
	}
}

func TestPareto_HeavierTailWithLowerShape(t *testing.T) {
	// Lower α → heavier tail → higher mean. Compare shape=1.5 vs shape=3.0.
	const min = 10
	sHeavy, _ := newPareto(config.Config{JitterMean: min, JitterShape: 1.5})
	sLight, _ := newPareto(config.Config{JitterMean: min, JitterShape: 3.0})

	mean := func(s Sampler) time.Duration {
		var total time.Duration
		for i := 0; i < samples; i++ {
			total += s.Sample()
		}
		return total / time.Duration(samples)
	}

	if mean(sHeavy) <= mean(sLight) {
		t.Fatal("expected lower shape (heavier tail) to produce higher mean")
	}
}
