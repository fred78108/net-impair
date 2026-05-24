package jitter

import (
	"testing"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

func TestNormal_NonNegative(t *testing.T) {
	// Large stddev relative to mean exercises the clamping path frequently.
	s, _ := newNormal(config.Config{JitterMean: 5, JitterStddev: 50})
	for i := 0; i < samples; i++ {
		if v := s.Sample(); v < 0 {
			t.Fatalf("negative sample: %v", v)
		}
	}
}

func TestNormal_Mean(t *testing.T) {
	const mean = 50
	s, _ := newNormal(config.Config{JitterMean: mean, JitterStddev: 10})
	var total time.Duration
	for i := 0; i < samples; i++ {
		total += s.Sample()
	}
	got := total / time.Duration(samples)
	want := time.Duration(mean) * time.Millisecond
	if !withinTolerance(got, want, 5*time.Millisecond) {
		t.Fatalf("mean %v not within 5ms of expected %v", got, want)
	}
}

func TestNormal_ZeroStddev(t *testing.T) {
	const mean = 30
	s, _ := newNormal(config.Config{JitterMean: mean, JitterStddev: 0})
	want := time.Duration(mean) * time.Millisecond
	for i := 0; i < 100; i++ {
		if v := s.Sample(); v != want {
			t.Fatalf("stddev=0: expected %v, got %v", want, v)
		}
	}
}
