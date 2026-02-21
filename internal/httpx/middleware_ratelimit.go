package httpx

import (
  "net/http"
  "sync"
  "time"

  "github.com/labstack/echo/v5"
  "golang.org/x/time/rate"
)

type ipLimiter struct {
  limiter  *rate.Limiter
  lastSeen time.Time
}

type rateLimiterStore struct {
  mu       sync.Mutex
  limiters map[string]*ipLimiter
  rps      rate.Limit
  burst    int
}

func newRateLimiterStore(rps float64, burst int) *rateLimiterStore {
  store := &rateLimiterStore{
    limiters: make(map[string]*ipLimiter),
    rps:      rate.Limit(rps),
    burst:    burst,
  }
  go store.cleanup()
  return store
}

func (s *rateLimiterStore) getLimiter(ip string) *rate.Limiter {
  s.mu.Lock()
  defer s.mu.Unlock()

  entry, exists := s.limiters[ip]
  if !exists {
    limiter := rate.NewLimiter(s.rps, s.burst)
    s.limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
    return limiter
  }
  entry.lastSeen = time.Now()
  return entry.limiter
}

func (s *rateLimiterStore) cleanup() {
  for {
    time.Sleep(time.Minute)
    s.mu.Lock()
    for ip, entry := range s.limiters {
      if time.Since(entry.lastSeen) > 3*time.Minute {
        delete(s.limiters, ip)
      }
    }
    s.mu.Unlock()
  }
}

func RateLimitMiddleware(rps float64, burst int) echo.MiddlewareFunc {
  store := newRateLimiterStore(rps, burst)
  return func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c *echo.Context) error {
      ip := c.RealIP()
      limiter := store.getLimiter(ip)
      if !limiter.Allow() {
        c.Response().Header().Set("Retry-After", "1")
        return writeErrorResponse(c, http.StatusTooManyRequests, ErrCodeRateLimited,
          "请求过于频繁，请稍后重试")
      }
      return next(c)
    }
  }
}
