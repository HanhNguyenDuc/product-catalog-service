package common

import "time"

// Ticker is a clock abstraction that allows injecting a deterministic time source in tests.
type Ticker interface {
	Now() time.Time
}

// RealTicker returns the current wall-clock time. Used in production.
type RealTicker struct{}

func (RealTicker) Now() time.Time { return time.Now() }

// NewRealTicker returns a production Ticker backed by time.Now().
func NewRealTicker() Ticker { return RealTicker{} }
