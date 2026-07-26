package image

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vendemas/imagekit/internal/cache"
	"github.com/vendemas/imagekit/internal/database"
	"github.com/vendemas/imagekit/internal/tenant"
	"github.com/vendemas/imagekit/internal/params"
	"github.com/vendemas/imagekit/internal/testutil"
)

func newTestProjectCache() *tenant.ProjectCache {
	pc := tenant.NewProjectCache(nil)
	pc.Set(&tenant.CachedProject{
		Project: &database.Project{
			Slug:           "testproj",
			BasePath:       "",
			DefaultQuality: 80,
			DefaultFormat:  "webp",
			MaxWidth:       2048,
			MaxHeight:      2048,
		},
		Provider: testutil.NewMockStorage(),
	})
	return pc
}

func TestHandlerImageNotFound(t *testing.T) {
	pc := newTestProjectCache()
	lru := cache.NewLRUCache(64, 3600)
	h := NewHandler(pc, lru, nil)

	r := chi.NewRouter()
	r.Get("/{slug}/{*}", h.Serve)

	rr := testutil.ServeHandler(r, "GET", "/testproj/nonexistent.jpg")
	testutil.AssertStatus(t, rr.Code, http.StatusNotFound)
}

func TestHandlerSlugNotFound(t *testing.T) {
	pc := newTestProjectCache()
	lru := cache.NewLRUCache(64, 3600)
	h := NewHandler(pc, lru, nil)

	r := chi.NewRouter()
	r.Get("/{slug}/{*}", h.Serve)

	rr := testutil.ServeHandler(r, "GET", "/noslug/file.jpg")
	testutil.AssertStatus(t, rr.Code, http.StatusNotFound)
}

func TestHandlerPathTraversal(t *testing.T) {
	pc := newTestProjectCache()
	lru := cache.NewLRUCache(64, 3600)
	h := NewHandler(pc, lru, nil)

	r := chi.NewRouter()
	r.Get("/{slug}/{*}", h.Serve)

	// URL-encoded path traversal
	rr := testutil.ServeHandler(r, "GET", "/testproj/..%2f..%2fetc%2fpasswd")
	testutil.AssertStatus(t, rr.Code, http.StatusForbidden)
}

func TestHandlerEmptyPath(t *testing.T) {
	pc := newTestProjectCache()
	lru := cache.NewLRUCache(64, 3600)
	h := NewHandler(pc, lru, nil)

	r := chi.NewRouter()
	r.Get("/{slug}/{*}", h.Serve)

	rr := testutil.ServeHandler(r, "GET", "/testproj/")
	testutil.AssertStatus(t, rr.Code, http.StatusNotFound)
}

func TestServeBytesSetsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	serveBytes(w, []byte("fake-image-data"), "image/jpeg", 86400)

	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("content-type: got %s, want image/jpeg", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "86400") {
		t.Errorf("cache-control missing max-age: got %s", cc)
	}
	if w.Header().Get("ETag") == "" {
		t.Error("ETag should be set")
	}
	if cl := w.Header().Get("Content-Length"); cl != "15" {
		t.Errorf("content-length: got %s, want 15", cl)
	}
}

func TestCacheKeyFormat(t *testing.T) {
	key := cacheKeyString("slug1", "path/img.jpg", params.Params{Width: 100})
	if !strings.HasPrefix(key, "slug1:path/img.jpg:") {
		t.Errorf("unexpected cache key format: %s", key)
	}
}

func TestValidateFilePath(t *testing.T) {
	if err := validateFilePath("normal/path.jpg"); err != nil {
		t.Errorf("expected no error for normal path, got %v", err)
	}
	if err := validateFilePath("../traversal.jpg"); err == nil {
		t.Error("expected error for path traversal")
	}
	if err := validateFilePath("/absolute/path.jpg"); err == nil {
		t.Error("expected error for absolute path")
	}
}

func TestDetectContentType(t *testing.T) {
	ct := detectContentType([]byte{0xFF, 0xD8, 0xFF}, "test.jpg")
	if !strings.HasPrefix(ct, "image/") {
		t.Errorf("expected image content type, got %s", ct)
	}
}

func TestSupportsFormat(t *testing.T) {
	if !supportsFormat("image/webp,image/avif,*/*", "webp") {
		t.Error("should detect webp support")
	}
	if !supportsFormat("image/avif,*/*", "avif") {
		t.Error("should detect avif support")
	}
	if supportsFormat("image/jpeg,*/*", "webp") {
		t.Error("should not detect webp when not in accept")
	}
	if supportsFormat("", "webp") {
		t.Error("empty accept should return false")
	}
}
