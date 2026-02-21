package search

import (
	"sync/atomic"
	"time"
)

// SearchActivityTracker 通过滑动时间窗口判断搜索服务是否处于活跃状态。
type SearchActivityTracker struct {
	lastActive atomic.Int64
	windowSec  int64
}

func NewSearchActivityTracker(windowSec int64) *SearchActivityTracker {
	return &SearchActivityTracker{windowSec: windowSec}
}

func (t *SearchActivityTracker) RecordActivity() {
	t.lastActive.Store(time.Now().Unix())
}

func (t *SearchActivityTracker) IsActive() bool {
	last := t.lastActive.Load()
	if last == 0 {
		return false
	}
	return time.Now().Unix()-last < t.windowSec
}
