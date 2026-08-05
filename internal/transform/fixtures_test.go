//go:build cgo

package transform

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

var (
	colWhite  = vips.ColorRGBA{R: 255, G: 255, B: 255, A: 255}
	colRed    = vips.ColorRGBA{R: 255, G: 0, B: 0, A: 255}
	colGreen  = vips.ColorRGBA{R: 0, G: 255, B: 0, A: 255}
	colBlue   = vips.ColorRGBA{R: 0, G: 0, B: 255, A: 255}
	colYellow = vips.ColorRGBA{R: 255, G: 255, B: 0, A: 255}
)

func makeSolidImage(t *testing.T, w, h int, bg color.RGBA) *image.RGBA {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	return img
}

// makeQuadrantsImage creates a w×h white image with four colored quadrants:
// red top-left, green top-right, blue bottom-left, yellow bottom-right.
// Each quadrant is a 1/4-size solid color with a 1/8-size gap from the edge.
func makeQuadrantsImage(t *testing.T, w, h int) *image.RGBA {
	t.Helper()
	img := makeSolidImage(t, w, h, color.RGBA{255, 255, 255, 255})

	qw := w / 4
	qh := h / 4
	gapX := w / 8
	gapY := h / 8

	draw.Draw(img, image.Rect(gapX, gapY, gapX+qw, gapY+qh), &image.Uniform{C: color.RGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(w-gapX-qw, gapY, w-gapX, gapY+qh), &image.Uniform{C: color.RGBA{G: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(gapX, h-gapY-qh, gapX+qw, h-gapY), &image.Uniform{C: color.RGBA{B: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(w-gapX-qw, h-gapY-qh, w-gapX, h-gapY), &image.Uniform{C: color.RGBA{R: 255, G: 255, A: 255}}, image.Point{}, draw.Src)

	return img
}

// makeTrimImage creates a w×h white image with a red rectangle in the middle,
// leaving a uniform border for the trim operation to detect.
func makeTrimImage(t *testing.T, w, h int) *image.RGBA {
	t.Helper()
	img := makeSolidImage(t, w, h, color.RGBA{255, 255, 255, 255})
	border := 20
	draw.Draw(img, image.Rect(border, border, w-border, h-border), &image.Uniform{C: color.RGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	return img
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// newQuadrantsPNG returns a PNG fixture with four colored quadrants.
func newQuadrantsPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	return encodePNG(t, makeQuadrantsImage(t, w, h))
}

// newTrimPNG returns a PNG fixture with a uniform border for trim tests.
func newTrimPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	return encodePNG(t, makeTrimImage(t, w, h))
}

// decode returns the decoded image ref; caller must Close it.
func decode(t *testing.T, buf []byte) *vips.ImageRef {
	t.Helper()
	img, err := vips.NewImageFromBuffer(buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return img
}

func dims(t *testing.T, buf []byte) (int, int) {
	t.Helper()
	img := decode(t, buf)
	defer img.Close()
	return img.Width(), img.Height()
}

// colorAt returns the RGBA value at (x, y), clamping to image bounds.
func colorAt(t *testing.T, buf []byte, x, y int) vips.ColorRGBA {
	t.Helper()
	img := decode(t, buf)
	defer img.Close()

	w, h := img.Width(), img.Height()
	if x >= w {
		x = w - 1
	}
	if y >= h {
		y = h - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	px, err := img.GetPoint(x, y)
	if err != nil {
		t.Fatalf("GetPoint(%d,%d): %v", x, y, err)
	}
	// A single-band (grayscale) image reports only one value; GetPoint pads
	// the remaining channels with garbage, so replicate the gray value.
	if img.Bands() == 1 {
		return vips.ColorRGBA{R: uint8(px[0]), G: uint8(px[0]), B: uint8(px[0]), A: 255}
	}
	c := vips.ColorRGBA{R: uint8(px[0]), G: uint8(px[1]), B: uint8(px[2]), A: 255}
	if len(px) > 3 {
		c.A = uint8(px[3])
	}
	return c
}

func near(a, b uint8, tol int) bool {
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func assertColor(t *testing.T, buf []byte, x, y int, want vips.ColorRGBA, tol int) {
	t.Helper()
	got := colorAt(t, buf, x, y)
	if !near(got.R, want.R, tol) || !near(got.G, want.G, tol) || !near(got.B, want.B, tol) {
		t.Errorf("color at (%d,%d) = %v, want %v (tol %d)", x, y, got, want, tol)
	}
}

func assertDims(t *testing.T, buf []byte, w, h int) {
	t.Helper()
	gw, gh := dims(t, buf)
	if gw != w || gh != h {
		t.Errorf("dims = %dx%d, want %dx%d", gw, gh, w, h)
	}
}

// saveOutput writes the raw output buffer to testdata/out/<name>.<ext> for
// visual inspection. ext defaults to "bin".
func saveOutput(t *testing.T, name, ext string, buf []byte) {
	t.Helper()
	if ext == "" {
		ext = "bin"
	}
	dir := filepath.Join("testdata", "out")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, name+"."+ext)
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// savePNG saves the output buffer re-encoded as PNG for easy visual inspection.
func savePNG(t *testing.T, name string, buf []byte) {
	t.Helper()
	img := decode(t, buf)
	defer img.Close()
	out, _, err := img.ExportPng(vips.NewPngExportParams())
	if err != nil {
		t.Fatalf("export png for %s: %v", name, err)
	}
	saveOutput(t, name, "png", out)
}
