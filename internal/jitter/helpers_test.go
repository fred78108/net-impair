package jitter

import "time"

const samples = 10_000

// withinTolerance reports whether got is within delta of want.
func withinTolerance(got, want, delta time.Duration) bool {
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	return diff <= delta
}
