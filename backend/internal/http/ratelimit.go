package http

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ipRateLimiter throttles requests per client IP using a token bucket, as a
// DoS mitigation. Stale per-IP limiters are evicted by a background sweeper so
// memory stays bounded even under a spray of distinct source IPs.
type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	rate    rate.Limit
	burst   int
	ttl     time.Duration
}

type ipBucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newIPRateLimiter builds a limiter allowing r tokens/sec with the given burst
// per client IP, and starts a sweeper that drops IPs idle longer than ttl.
func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		buckets: make(map[string]*ipBucket),
		rate:    r,
		burst:   burst,
		ttl:     10 * time.Minute,
	}
	go l.sweep()
	return l
}

func (l *ipRateLimiter) sweep() {
	for range time.Tick(l.ttl) {
		l.mu.Lock()
		for ip, b := range l.buckets {
			if time.Since(b.lastSeen) > l.ttl {
				delete(l.buckets, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	b, ok := l.buckets[ip]
	if !ok {
		b = &ipBucket{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.buckets[ip] = b
	}
	b.lastSeen = time.Now()
	l.mu.Unlock()
	return b.limiter.Allow()
}

// middleware rejects (429) requests from an IP that exceeds the configured rate.
func (l *ipRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "muitas requisições, tente novamente em instantes"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP returns the host portion of RemoteAddr. middleware.RealIP has
// already normalized RemoteAddr from X-Forwarded-For/X-Real-IP when the request
// arrived through a trusted proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// maxBodyBytes caps the request body size to guard against large-payload
// memory-exhaustion DoS. Handlers reading an oversized body get an error from
// the wrapped reader.
func maxBodyBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
