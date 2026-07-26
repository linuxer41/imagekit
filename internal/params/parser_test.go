package params

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	p := ParseTR("")
	if p.HasParams() {
		t.Error("empty string should have no params")
	}
}

func TestParseWidth(t *testing.T) {
	p := ParseTR("w-800")
	if !p.HasParams() {
		t.Fatal("expected params")
	}
	if p.Width != 800 {
		t.Errorf("width: got %d, want 800", p.Width)
	}
}

func TestParseMultiple(t *testing.T) {
	p := ParseTR("w-800,h-600,q-80,f-webp")
	if p.Width != 800 {
		t.Errorf("width: got %d, want 800", p.Width)
	}
	if p.Height != 600 {
		t.Errorf("height: got %d, want 600", p.Height)
	}
	if p.Quality != 80 {
		t.Errorf("quality: got %d, want 80", p.Quality)
	}
	if p.Format != "webp" {
		t.Errorf("format: got %s, want webp", p.Format)
	}
}

func TestParseAspectRatio(t *testing.T) {
	p := ParseTR("ar-4-3")
	if !p.HasParams() {
		t.Fatal("expected params")
	}
	if p.Aspect != "4-3" {
		t.Errorf("aspect: got %s, want 4-3", p.Aspect)
	}
}

func TestParseBlur(t *testing.T) {
	p := ParseTR("bl-2.5")
	if p.Blur != 2.5 {
		t.Errorf("blur: got %f, want 2.5", p.Blur)
	}
}

func TestParseAllParams(t *testing.T) {
	raw := "w-300,h-200,ar-16-9,cm-force,q-75,f-avif," +
		"bg-00000000,rt-180,fo-face,bl-1,e-sharpen," +
		"b-5_FF0000,pr-true,lo-true,l-10,pg-1"

	p := ParseTR(raw)
	if !p.HasParams() {
		t.Fatal("expected params")
	}
	if p.Width != 300 {
		t.Errorf("width: got %d", p.Width)
	}
	if p.Format != "avif" {
		t.Errorf("format: got %s", p.Format)
	}
	if p.Rotation != 180 {
		t.Errorf("rotation: got %d", p.Rotation)
	}
	if !p.Progressive {
		t.Error("progressive should be true")
	}
	if !p.Lossless {
		t.Error("lossless should be true")
	}
	if p.Page != 1 {
		t.Errorf("page: got %d", p.Page)
	}
}

func TestParseInvalidParams(t *testing.T) {
	raw := "w-abc,h-0,q--5,rt-abc,bl-invalid"
	p := ParseTR(raw)
	if p.HasParams() {
		t.Error("invalid params should not set hasParams")
	}
}
