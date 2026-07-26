package testutil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func AssertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status: got %d, want %d", got, want)
	}
}

func AssertBody(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("body does not contain expected: got %q, want substr %q", got, want)
	}
}

func DecodeResponse(rr *httptest.ResponseRecorder) string {
	body, _ := io.ReadAll(rr.Body)
	return string(body)
}

type errNotFound string

func (e errNotFound) Error() string { return fmt.Sprintf("not found: %s", string(e)) }

func (e errNotFound) Is(target error) bool {
	return strings.Contains(target.Error(), "not found")
}

func IsNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func ServeHandler(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func ServeHandlerWithHeader(h http.Handler, method, path, key, val string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set(key, val)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func ServeHandlerWithBody(h http.Handler, method, path, bodyStr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, io.NopCloser(strings.NewReader(bodyStr)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}
