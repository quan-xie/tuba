package backoff

import (
	"github.com/quan-xie/tuba/util/xtime"
	"math"
	"time"
)

// Backoff interface defines contract for backoff strategies
type Backoff interface {
	Next(retry int) time.Duration
}

type constantBackoff struct {
	backoffInterval xtime.Duration
}

// NewConstantBackoff returns an instance of ConstantBackoff
func NewConstantBackoff(backoffInterval xtime.Duration) Backoff {
	return &constantBackoff{backoffInterval: backoffInterval}
}

// Next returns next time for retrying operation with constant strategy
func (cb *constantBackoff) Next(retry int) time.Duration {
	if retry <= 0 {
		return 0 * time.Millisecond
	}

	return time.Duration(cb.backoffInterval) * 1 << uint(retry)
}

type exponentialBackoff struct {
	exponentFactor float64
	initialTimeout float64
	maxTimeout     float64
}

// NewExponentialBackoff returns an instance of ExponentialBackoff
func NewExponentialBackoff(initialTimeout, maxTimeout time.Duration, exponentFactor float64) Backoff {
	return &exponentialBackoff{
		exponentFactor: exponentFactor,
		initialTimeout: float64(initialTimeout / time.Millisecond),
		maxTimeout:     float64(maxTimeout / time.Millisecond),
	}
}

// Next returns next time for retrying operation with exponential strategy
func (eb *exponentialBackoff) Next(retry int) time.Duration {
	if retry <= 0 {
		return 0 * time.Millisecond
	}

	return time.Duration(math.Min(eb.initialTimeout+math.Pow(eb.exponentFactor, float64(retry)), eb.maxTimeout)) * time.Millisecond
}
