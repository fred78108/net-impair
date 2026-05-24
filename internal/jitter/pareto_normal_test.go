package jitter

import (
	"testing"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

func TestParetoNormal_NonNegative(t *testing.T) {
	s, _ := newParetoNormal(config.Config{
		JitterMean: 20, JitterStddev: 5, JitterShape: 2, JitterMix: 0.5,
	})
	for i := 0; i < samples; i++ {
		if v := s.Sample(); v < 0 {
			t.Fatalf("negative sample: %v", v)
		}
	}
}

func TestParetoNormal_PureParetoMix(t *testing.T) {
	// mix=1 → pure Pareto; all samples must be >= min.
	const min = 10
	s, _ := newParetoNormal(config.Config{
		JitterMean: min, JitterStddev: 5, JitterShape: 2, JitterMix: 1.0,
	})
	minDur := time.Duration(min) * time.Millisecond
	for i := 0; i < samples; i++ {
		if v := s.Sample(); v < minDur {
			t.Fatalf("pure-Pareto mix: sample %v below min %v", v, minDur)
		}
	}
}

func TestParetoNormal_PureNormalMix(t *testing.T) {
	// mix=0 → pure Normal; no negative samples, mean within tolerance.
	const mean = 50
	s, _ := newParetoNormal(config.Config{
		JitterMean: mean, JitterStddev: 10, JitterShape: 2, JitterMix: 0.0,
	})
	var total time.Duration
	for i := 0; i < samples; i++ {
		v := s.Sample()
		if v < 0 {
			t.Fatalf("pure-Normal mix: negative sample %v", v)
		}
		total += v
	}
	got := total / time.Duration(samples)
	want := time.Duration(mean) * time.Millisecond
	if !withinTolerance(got, want, 5*time.Millisecond) {
		t.Fatalf("pure-Normal mix mean %v not within 5ms of expected %v", got, want)
	}
}
