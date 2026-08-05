//go:build cgo

package transform

import (
	"os"
	"strings"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/vendemas/imagekit/internal/params"
)

func TestMain(m *testing.M) {
	if err := vips.Startup(&vips.Config{}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type pixelCheck struct {
	x, y int
	want vips.ColorRGBA
	tol  int
}

type trCase struct {
	name    string
	raw     string
	fixture []byte // nil → default quadrants 200x150
	wantCT  string
	wantW   int
	wantH   int
	pixels  []pixelCheck
	// magic is a list of byte sequences to find in the output, each at an
	// explicit offset (used to check file signatures, e.g. "RIFF"@0+"WEBP"@8).
	magic []magicAt
	// equalChannels positions that must have R==G==B (grayscale).
	equalChannels [][2]int
	// skipOn contains substrings; if any appears in the error the test is skipped
	// (used for codecs not compiled into a given libvips build).
	skipOn []string
	// skipName is the message logged when the test is skipped.
	skipName string
}

type magicAt struct {
	offset int
	bytes  []byte
}

func TestProcessImageCases(t *testing.T) {
	quad := newQuadrantsPNG(t, 200, 150)
	trim := newTrimPNG(t, 200, 150)

	red := colRed
	white := colWhite

	cases := []trCase{
		{name: "resize-w", raw: "w-800", wantCT: "image/png", wantW: 800, wantH: 600,
			pixels: []pixelCheck{{200, 144, red, 15}}},
		{name: "resize-h", raw: "h-600", wantCT: "image/png", wantW: 800, wantH: 600,
			pixels: []pixelCheck{{200, 144, red, 15}}},
		{name: "resize-w-h", raw: "w-400,h-300", wantCT: "image/png", wantW: 400, wantH: 300,
			pixels: []pixelCheck{{100, 72, red, 15}}},
		{name: "force-crop", raw: "w-300,h-300,cm-force", wantCT: "image/png", wantW: 300, wantH: 300,
			pixels: []pixelCheck{{50, 70, red, 15}}},
		{name: "at-max-no-upscale", raw: "w-400,h-300,cm-at_max", wantCT: "image/png", wantW: 200, wantH: 150,
			pixels: []pixelCheck{{50, 36, red, 5}}},
		{name: "at-max-shrink", raw: "w-100,h-100,cm-at_max", wantCT: "image/png", wantW: 100, wantH: 75},
		{name: "maintain-ratio", raw: "w-400,h-300,cm-maintain_ratio", wantCT: "image/png", wantW: 400, wantH: 300},
		{name: "at-least", raw: "w-400,h-300,cm-at_least", wantCT: "image/png", wantW: 400, wantH: 300},
		{name: "max-size-enum", raw: "w-400,h-300,cm-max_size_enum", wantCT: "image/png", wantW: 400, wantH: 300},
		{name: "pad-resize-fit", raw: "w-400,h-300,cm-pad_resize", wantCT: "image/png", wantW: 400, wantH: 300,
			pixels: []pixelCheck{{100, 72, red, 15}}},
		{name: "pad-resize-red-bg", raw: "w-400,h-400,cm-pad_resize,bg-FF0000", wantCT: "image/png", wantW: 400, wantH: 400,
			pixels: []pixelCheck{{0, 0, red, 5}, {399, 399, red, 5}, {100, 120, red, 15}}},
		{name: "pad-resize-width-only", raw: "w-400,cm-pad_resize", wantCT: "image/png", wantW: 400, wantH: 300},
		{name: "aspect-4-3", raw: "ar-4-3", wantCT: "image/png", wantW: 200, wantH: 150,
			pixels: []pixelCheck{{50, 36, red, 5}}},
		{name: "aspect-16-9", raw: "ar-16-9", wantCT: "image/png", wantW: 200, wantH: 112},
		{name: "aspect-1-1", raw: "ar-1-1", wantCT: "image/png", wantW: 150, wantH: 150},
		{name: "rotate-90", raw: "rt-90", wantCT: "image/png", wantW: 150, wantH: 200,
			pixels: []pixelCheck{{0, 0, white, 20}}},
		{name: "rotate-90-bg", raw: "rt-90,bg-FF0000", wantCT: "image/png", wantW: 150, wantH: 200,
			pixels: []pixelCheck{{0, 0, red, 20}}},
		{name: "rotate-180", raw: "rt-180", wantCT: "image/png", wantW: 200, wantH: 150,
			pixels: []pixelCheck{{150, 114, red, 40}}},
		{name: "rotate-270", raw: "rt-270", wantCT: "image/png", wantW: 150, wantH: 200},
		{name: "format-jpg", raw: "f-jpg", wantCT: "image/jpeg", wantW: 200, wantH: 150,
			magic: []magicAt{{0, []byte{0xFF, 0xD8, 0xFF}}}},
		{name: "format-png", raw: "f-png", wantCT: "image/png", wantW: 200, wantH: 150,
			magic: []magicAt{{0, []byte{0x89, 'P', 'N', 'G'}}}},
		{name: "format-webp", raw: "f-webp", wantCT: "image/webp", wantW: 200, wantH: 150,
			magic: []magicAt{{0, []byte{'R', 'I', 'F', 'F'}}, {8, []byte{'W', 'E', 'B', 'P'}}}},
		{name: "format-avif", raw: "f-avif", wantCT: "image/avif", wantW: 200, wantH: 150,
			magic:    []magicAt{{4, []byte{'f', 't', 'y', 'p'}}},
			skipOn:   []string{"not found", "unsupported", "unable to load", "codec", "heif", "avif", "no such operation", "unable to save", "save", "export"},
			skipName: "avif codec not available"},
		{name: "format-gif", raw: "f-gif", wantCT: "image/gif", wantW: 200, wantH: 150,
			magic: []magicAt{{0, []byte{'G', 'I', 'F', '8'}}}},
		{name: "format-auto", raw: "", wantCT: "image/png", wantW: 200, wantH: 150,
			pixels: []pixelCheck{{50, 36, red, 5}}},
		{name: "quality-jpg", raw: "q-80,f-jpg", wantCT: "image/jpeg", wantW: 200, wantH: 150,
			magic: []magicAt{{0, []byte{0xFF, 0xD8, 0xFF}}}},
		{name: "grayscale", raw: "e-grayscale", wantCT: "image/png", wantW: 200, wantH: 150,
			equalChannels: [][2]int{{50, 36}}},
		{name: "sharpen", raw: "e-sharpen", wantCT: "image/png", wantW: 200, wantH: 150},
		{name: "contrast-ignored", raw: "e-contrast", wantCT: "image/png", wantW: 200, wantH: 150},
		{name: "bright-ignored", raw: "e-bright", wantCT: "image/png", wantW: 200, wantH: 150},
		{name: "blur", raw: "bl-3.0", wantCT: "image/png", wantW: 200, wantH: 150},
		{name: "trim", raw: "t-10", fixture: trim, wantCT: "image/png", wantW: 160, wantH: 110,
			pixels: []pixelCheck{{0, 0, red, 5}}},
		{name: "progressive-jpg", raw: "pr-true,f-jpg", wantCT: "image/jpeg", wantW: 200, wantH: 150},
		{name: "lossless-webp", raw: "lo-true,f-webp", wantCT: "image/webp", wantW: 200, wantH: 150},
		{name: "metadata-png", raw: "md-true,f-png", wantCT: "image/png", wantW: 200, wantH: 150},
		{name: "ignored-params", raw: "fo-face,b-5_FF0000,l-10,pg-1", wantCT: "image/png", wantW: 200, wantH: 150,
			pixels: []pixelCheck{{50, 36, red, 5}}},
		{name: "combined", raw: "ar-4-3,w-800,f-webp,q-80", wantCT: "image/webp", wantW: 800, wantH: 600},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			src := c.fixture
			if src == nil {
				src = quad
			}
			p := params.ParseTR(c.raw)
			out, ct, err := ProcessImage(src, p)
			if err != nil {
				if skipReason(c.skipOn, err) {
					t.Skipf("%s: %v", c.skipName, err)
				}
				t.Fatalf("ProcessImage(%q): %v", c.raw, err)
			}
			if ct != c.wantCT {
				t.Errorf("content-type = %q, want %q", ct, c.wantCT)
			}
			assertDims(t, out, c.wantW, c.wantH)
			for _, pc := range c.pixels {
				assertColor(t, out, pc.x, pc.y, pc.want, pc.tol)
			}
			for _, pos := range c.equalChannels {
				assertEqualChannels(t, out, pos[0], pos[1])
			}
			assertMagic(t, out, c.magic)
			savePNG(t, c.name, out)
		})
	}
}

func skipReason(substrings []string, err error) bool {
	if len(substrings) == 0 {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, s := range substrings {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func assertEqualChannels(t *testing.T, buf []byte, x, y int) {
	t.Helper()
	c := colorAt(t, buf, x, y)
	if !near(c.R, c.G, 20) || !near(c.G, c.B, 20) {
		t.Errorf("pixel (%d,%d) not grayscale: %v", x, y, c)
	}
}

func assertMagic(t *testing.T, buf []byte, magic []magicAt) {
	t.Helper()
	for _, m := range magic {
		if len(buf) < m.offset+len(m.bytes) {
			t.Errorf("output too short (%d bytes) for magic at offset %d", len(buf), m.offset)
			continue
		}
		for j, b := range m.bytes {
			if buf[m.offset+j] != b {
				t.Errorf("magic byte at offset %d = %#x, want %#x", m.offset+j, buf[m.offset+j], b)
			}
		}
	}
}
