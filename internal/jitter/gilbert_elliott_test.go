package jitter

import (
	"sync"
	"testing"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

func TestGilbertElliott_OnlyValidDelays(t *testing.T) {
	const good, bad = 10, 100
	s, _ := newGilbertElliott(config.Config{
		GoodDelay: good, BadDelay: bad, PGoodToBad: 0.2, PBadToGood: 0.4,
	})
	goodDur := time.Duration(good) * time.Millisecond
	badDur := time.Duration(bad) * time.Millisecond
	for i := 0; i < samples; i++ {
		v := s.Sample()
		if v != goodDur && v != badDur {
			t.Fatalf("unexpected delay %v; want %v or %v", v, goodDur, badDur)
		}
	}
}

func TestGilbertElliott_AlwaysBad(t *testing.T) {
	// pGoodToBad=1, pBadToGood=0: first Sample transitions to bad, all subsequent stay bad.
	s, _ := newGilbertElliott(config.Config{
		GoodDelay: 10, BadDelay: 100, PGoodToBad: 1.0, PBadToGood: 0.0,
	})
	badDur := 100 * time.Millisecond
	s.Sample() // first call transitions good → bad
	for i := 0; i < 100; i++ {
		if v := s.Sample(); v != badDur {
			t.Fatalf("expected bad delay %v, got %v", badDur, v)
		}
	}
}

func TestGilbertElliott_AlwaysGood(t *testing.T) {
	// pGoodToBad=0: never leaves good state.
	s, _ := newGilbertElliott(config.Config{
		GoodDelay: 10, BadDelay: 100, PGoodToBad: 0.0, PBadToGood: 1.0,
	})
	goodDur := 10 * time.Millisecond
	for i := 0; i < 100; i++ {
		if v := s.Sample(); v != goodDur {
			t.Fatalf("expected good delay %v, got %v", goodDur, v)
		}
	}
}

func TestGilbertElliott_BothStatesReached(t *testing.T) {
	// High transition probabilities ensure both states are visited.
	s, _ := newGilbertElliott(config.Config{
		GoodDelay: 10, BadDelay: 100, PGoodToBad: 0.5, PBadToGood: 0.5,
	})
	goodDur := 10 * time.Millisecond
	badDur := 100 * time.Millisecond
	seenGood, seenBad := false, false
	for i := 0; i < samples; i++ {
		switch s.Sample() {
		case goodDur:
			seenGood = true
		case badDur:
			seenBad = true
		}
		if seenGood && seenBad {
			return
		}
	}
	t.Fatalf("after %d samples: seenGood=%v seenBad=%v", samples, seenGood, seenBad)
}

func TestGilbertElliott_ConcurrentSafety(t *testing.T) {
	s, _ := newGilbertElliott(config.Config{
		GoodDelay: 10, BadDelay: 100, PGoodToBad: 0.3, PBadToGood: 0.3,
	})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1_000; j++ {
				s.Sample()
			}
		}()
	}
	wg.Wait()
}
