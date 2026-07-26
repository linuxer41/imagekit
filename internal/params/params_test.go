package params

import (
	"testing"
)

func TestCacheKeyDiffers(t *testing.T) {
	a := Params{Width: 100, Height: 200, Quality: 80}
	b := Params{Width: 100, Height: 200, Quality: 90}
	if a.CacheKey() == b.CacheKey() {
		t.Error("different params should have different cache keys")
	}
}

func TestCacheKeySame(t *testing.T) {
	a := Params{Width: 100, Height: 200, Format: "webp"}
	b := Params{Width: 100, Height: 200, Format: "webp"}
	if a.CacheKey() != b.CacheKey() {
		t.Error("same params should have same cache key")
	}
}

func TestHasParamsDefault(t *testing.T) {
	p := Params{}
	if p.HasParams() {
		t.Error("default params should not have hasParams")
	}
}

func TestHasParamsWidth(t *testing.T) {
	p := Params{Width: 100}
	if p.HasParams() {
		t.Error("Width alone should not set hasParams")
	}
}
