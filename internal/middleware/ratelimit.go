package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMutations allows each client IP at most limit requests per window
// to anything but GET, HEAD, and OPTIONS. Every authentication endpoint is a
// POST, so this guards sign-in as well as state changes. Counts live in
// memory, so each API instance limits on its own.
func RateLimitMutations(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &fixedWindow{limit: limit, window: window, counts: map[string]int{}}
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		if wait, ok := limiter.allow(c.ClientIP(), time.Now()); !ok {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
			return
		}
		c.Next()
	}
}

// fixedWindow counts requests per key in windows shared by every key.
// Dropping all counts when a window ends bounds memory by the clients seen in
// one window.
type fixedWindow struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	start  time.Time
	counts map[string]int
}

// allow records a request and reports whether it fits, or else how long
// until the window resets.
func (f *fixedWindow) allow(key string, now time.Time) (time.Duration, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if now.Sub(f.start) >= f.window {
		f.start = now
		clear(f.counts)
	}
	if f.counts[key] >= f.limit {
		return f.start.Add(f.window).Sub(now), false
	}
	f.counts[key]++
	return 0, true
}
