package jitter

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/fred78108/net-impair/internal/config"
)

// gilbertElliottSampler implements a two-state Markov chain (Good/Bad) that
// produces correlated, bursty delays. It is the only stateful model: the current
// state is carried across successive [gilbertElliottSampler.Sample] calls.
// A sync.Mutex serialises state transitions so the sampler is safe for concurrent use.
// Registered under the model name "gilbert_elliott".
type gilbertElliottSampler struct {
	mu         sync.Mutex
	inBad      bool
	goodDelay  time.Duration
	badDelay   time.Duration
	pGoodToBad float64
	pBadToGood float64
}

// newGilbertElliott constructs a gilbertElliottSampler from cfg.GoodDelay,
// cfg.BadDelay, cfg.PGoodToBad, and cfg.PBadToGood.
// The chain starts in the good state.
func newGilbertElliott(cfg config.Config) (Sampler, error) {
	return &gilbertElliottSampler{
		goodDelay:  time.Duration(cfg.GoodDelay) * time.Millisecond,
		badDelay:   time.Duration(cfg.BadDelay) * time.Millisecond,
		pGoodToBad: cfg.PGoodToBad,
		pBadToGood: cfg.PBadToGood,
	}, nil
}

// Sample advances the Markov chain by one step and returns the delay for the
// resulting state: goodDelay in the good state, badDelay in the bad state.
func (s *gilbertElliottSampler) Sample() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inBad {
		if rand.Float64() < s.pBadToGood {
			s.inBad = false
		}
	} else {
		if rand.Float64() < s.pGoodToBad {
			s.inBad = true
		}
	}

	if s.inBad {
		return s.badDelay
	}
	return s.goodDelay
}
