package jitter

import (
	"testing"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

func TestUniform_ZeroMax(t *testing.T) {
	s, _ := newUniform(config.Config{JitterMax: 0})
	for i := 0; i < 100; i++ {
		if v := s.Sample(); v != 0 {
			t.Fatalf("max=0: expected 0, got %v", v)
		}
	}
}

func TestUniform_Range(t *testing.T) {
	const max = 100
	s, _ := newUniform(config.Config{JitterMax: max})
	maxDur := time.Duration(max) * time.Millisecond
	for i := 0; i < samples; i++ {
		v := s.Sample()
		if v < 0 || v > maxDur {
			t.Fatalf("sample %v outside [0, %v]", v, maxDur)
		}
	}
}

func TestUniform_Mean(t *testing.T) {
	const max = 100
	s, _ := newUniform(config.Config{JitterMax: max})
	var total time.Duration
	for i := 0; i < samples; i++ {
		total += s.Sample()
	}
	got := total / time.Duration(samples)
	want := time.Duration(max/2) * time.Millisecond
	if !withinTolerance(got, want, 5*time.Millisecond) {
		t.Fatalf("mean %v not within 5ms of expected %v", got, want)
	}
}
