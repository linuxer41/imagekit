package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowedOrigin(t *testing.T) {
	cors := NewCORS("https://app.vendemas.icu")

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://app.vendemas.icu")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://app.vendemas.icu" {
		t.Errorf("expected CORS header, got %s", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	cors := NewCORS("https://app.vendemas.icu")

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("disallowed origin should not get CORS header")
	}
}

func TestCORSWildcard(t *testing.T) {
	cors := NewCORS("*")

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://anything.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://anything.com" {
		t.Errorf("wildcard should echo origin, got %s", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSPreflight(t *testing.T) {
	cors := NewCORS("*")

	handler := cors.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://test.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("preflight should return 204, got %d", rr.Code)
	}
}
