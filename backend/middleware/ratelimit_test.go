package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimitAuth_AllowsUnderLimit(t *testing.T) {
	l := newIPLimiter(5, 5)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.get(realIP(r)).Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	for i := range 5 {
		req := httptest.NewRequest("POST", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimitAuth_BlocksOverLimit(t *testing.T) {
	l := newIPLimiter(5, 3)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.get(realIP(r)).Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	ip := "10.0.0.1:9999"
	blocked := 0
	for range 10 {
		req := httptest.NewRequest("POST", "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			blocked++
		}
	}
	if blocked == 0 {
		t.Fatal("expected at least one blocked request")
	}
}

func TestRateLimitAuth_IsolatesIPs(t *testing.T) {
	l := newIPLimiter(1, 1)
	exhausted := false
	for range 5 {
		req := httptest.NewRequest("POST", "/", nil)
		req.RemoteAddr = "1.1.1.1:1"
		if !l.get(realIP(req)).Allow() {
			exhausted = true
			break
		}
	}
	if !exhausted {
		t.Fatal("expected IP 1.1.1.1 to be rate limited")
	}
	req2 := httptest.NewRequest("POST", "/", nil)
	req2.RemoteAddr = "2.2.2.2:1"
	if !l.get(realIP(req2)).Allow() {
		t.Fatal("different IP should not be rate limited")
	}
}

func TestRealIP_XForwardedFor_WithTrustedProxy(t *testing.T) {
	os.Setenv("TRUSTED_PROXY", "192.0.2.1")
	defer os.Unsetenv("TRUSTED_PROXY")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if got := realIP(req); got != "1.2.3.4" {
		t.Fatalf("want 1.2.3.4, got %s", got)
	}
}

func TestRealIP_XForwardedFor_UntrustedProxy(t *testing.T) {
	os.Unsetenv("TRUSTED_PROXY")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	if got := realIP(req); got != "1.2.3.4" {
		t.Fatalf("want RemoteAddr 1.2.3.4, got %s", got)
	}
}

func TestRealIP_XRealIP_WithTrustedProxy(t *testing.T) {
	os.Setenv("TRUSTED_PROXY", "192.0.2.1")
	defer os.Unsetenv("TRUSTED_PROXY")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("X-Real-IP", "9.9.9.9")
	if got := realIP(req); got != "9.9.9.9" {
		t.Fatalf("want 9.9.9.9, got %s", got)
	}
}

func TestRealIP_RemoteAddr(t *testing.T) {
	os.Unsetenv("TRUSTED_PROXY")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:5678"
	if got := realIP(req); got != "127.0.0.1" {
		t.Fatalf("want 127.0.0.1, got %s", got)
	}
}
