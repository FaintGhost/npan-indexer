package indexer

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// ActivityChecker 用于判断搜索是否处于活跃状态。
type ActivityChecker interface {
	IsActive() bool
}

type RequestLimiter struct {
	concurrency chan struct{}
	limiter     *rate.Limiter
	baseRate    rate.Limit
	checker     ActivityChecker
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
		baseRate:    baseRate,
	}
}

func (l *RequestLimiter) SetActivityChecker(checker ActivityChecker) {
	l.checker = checker
}

func (l *RequestLimiter) adjustRate() {
	if l.checker == nil {
		return
	}
	if l.checker.IsActive() {
		l.limiter.SetLimit(l.baseRate / 2)
	} else {
		l.limiter.SetLimit(l.baseRate)
	}
}

func (l *RequestLimiter) Schedule(ctx context.Context, fn func() error) error {
	l.concurrency <- struct{}{}
	defer func() { <-l.concurrency }()

	l.adjustRate()

	if err := l.limiter.Wait(ctx); err != nil {
		return err
	}

	return fn()
}
