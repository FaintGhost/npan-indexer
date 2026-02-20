package indexer

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type RequestLimiter struct {
	concurrency chan struct{}
	limiter     *rate.Limiter
}

func NewRequestLimiter(maxConcurrent int, minTimeMS int) *RequestLimiter {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	if minTimeMS < 0 {
		minTimeMS = 0
	}

	minInterval := time.Duration(minTimeMS) * time.Millisecond
	baseRate := rate.Inf
	burst := 1
	if minInterval > 0 {
		baseRate = rate.Every(minInterval)
	}

	return &RequestLimiter{
		concurrency: make(chan struct{}, maxConcurrent),
		limiter:     rate.NewLimiter(baseRate, burst),
	}
}

func (l *RequestLimiter) Schedule(ctx context.Context, fn func() error) error {
	l.concurrency <- struct{}{}
	defer func() { <-l.concurrency }()

	if err := l.limiter.Wait(ctx); err != nil {
		return err
	}

	return fn()
}
