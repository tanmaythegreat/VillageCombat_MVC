package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu       sync.Mutex
	visitors = make(map[string]*limiterEntry)
)

func getLimiter(key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	e, exists := visitors[key]
	if !exists {
		// e.g. 5 requests, refilling 1 every 6s -> ~10/min sustained, burst 5
		l := rate.NewLimiter(rate.Every(6*time.Second), 5)
		visitors[key] = &limiterEntry{limiter: l, lastSeen: time.Now()}
		return l
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func Cleanup() {
	mu.Lock()
	defer mu.Unlock()
	for k, e := range visitors {
		if time.Since(e.lastSeen) > 10*time.Minute {
			delete(visitors, k)
		}
	}
}

func RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		if !getLimiter(host).Allow() {
			http.Error(w, "Too many requests, slow down", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
