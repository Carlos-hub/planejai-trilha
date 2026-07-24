package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiterBurstThenBlock(t *testing.T) {
	// rate.Every large + burst 3: first 3 allowed, 4th denied for the same IP.
	l := newIPRateLimiter(rate.Every(time.Hour), 3)
	ip := "10.0.0.1"
	for i := 0; i < 3; i++ {
		if !l.allow(ip) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if l.allow(ip) {
		t.Fatal("4th request from same IP should be blocked")
	}
	// A different IP has its own bucket and is not affected.
	if !l.allow("10.0.0.2") {
		t.Fatal("first request from a fresh IP should be allowed")
	}
}

func TestRateLimitMiddleware429(t *testing.T) {
	l := newIPRateLimiter(rate.Every(time.Hour), 1)
	h := l.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "192.0.2.5:12345"

	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request = %d, want 200", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request = %d, want 429", w2.Code)
	}
}
