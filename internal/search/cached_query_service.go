package search

import (
  "fmt"
  "strings"
  "time"

  "github.com/hashicorp/golang-lru/v2/expirable"

  "npan/internal/models"
)

// CachedQueryService 是 Searcher 的缓存装饰器，使用 LRU + TTL 策略。
type CachedQueryService struct {
  inner   Searcher
  cache   *expirable.LRU[string, QueryResult]
  tracker *SearchActivityTracker
}

// NewCachedQueryService 创建一个带 LRU 缓存的搜索服务装饰器。
// tracker 可选，传入后每次 Query 会记录搜索活动。
func NewCachedQueryService(inner Searcher, capacity int, ttl time.Duration, tracker *SearchActivityTracker) *CachedQueryService {
  return &CachedQueryService{
    inner:   inner,
    cache:   expirable.NewLRU[string, QueryResult](capacity, nil, ttl),
    tracker: tracker,
  }
}

// cacheKey 将搜索参数序列化为确定性的缓存键。
func cacheKey(p models.LocalSearchParams) string {
  var b strings.Builder
  fmt.Fprintf(&b, "%s|%s|%d|%d", p.Query, p.Type, p.Page, p.PageSize)

  if p.ParentID != nil {
    fmt.Fprintf(&b, "|p%d", *p.ParentID)
  }
  if p.UpdatedAfter != nil {
    fmt.Fprintf(&b, "|a%d", *p.UpdatedAfter)
  }
  if p.UpdatedBefore != nil {
    fmt.Fprintf(&b, "|b%d", *p.UpdatedBefore)
  }
  if p.IncludeDeleted {
    b.WriteString("|d")
  }

  return b.String()
}

func (s *CachedQueryService) Query(params models.LocalSearchParams) (QueryResult, error) {
  if s.tracker != nil {
    s.tracker.RecordActivity()
  }

  key := cacheKey(params)

  if cached, ok := s.cache.Get(key); ok {
    return cached, nil
  }

  result, err := s.inner.Query(params)
  if err != nil {
    return QueryResult{}, err
  }

  s.cache.Add(key, result)
  return result, nil
}

func (s *CachedQueryService) Ping() error {
  return s.inner.Ping()
}
