//go:build cgo

package transform

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/vendemas/imagekit/internal/params"
)

func ProcessImage(input []byte, p params.Params) ([]byte, string, error) {
	img, err := vips.NewImageFromBuffer(input)
	if err != nil {
		return nil, "", fmt.Errorf("decode: %w", err)
	}

	if p.Trim > 0 {
		// govips FindTrim dereferences the background color unconditionally,
		// so sample the top-left corner instead of passing nil.
		px, err := img.GetPoint(0, 0)
		if err != nil {
			return nil, "", fmt.Errorf("sample trim background: %w", err)
		}
		bg := &vips.Color{R: uint8(px[0]), G: uint8(px[1]), B: uint8(px[2])}
		left, top, width, height, err := img.FindTrim(float64(p.Trim), bg)
		if err != nil {
			return nil, "", fmt.Errorf("find trim: %w", err)
		}
		if width > 0 && height > 0 {
			if err := img.ExtractArea(left, top, width, height); err != nil {
				return nil, "", fmt.Errorf("trim: %w", err)
			}
		}
	}

	if p.Rotation != 0 {
		bg := vips.ColorRGBA{R: 255, G: 255, B: 255, A: 255}
		if p.BgColor != "" {
			bg = parseBGColor(p.BgColor)
		}
		err := img.Similarity(1.0, float64(p.Rotation), &bg, 0.5, 0.5, 0.5, 0.5)
		if err != nil {
			return nil, "", fmt.Errorf("rotate: %w", err)
		}
	}

	if p.Width > 0 || p.Height > 0 || p.Aspect != "" {
		err = applyResize(img, p)
		if err != nil {
			return nil, "", fmt.Errorf("resize: %w", err)
		}
	}

	if p.Effect == params.EffectGrayscale {
		if err := img.ToColorSpace(vips.InterpretationBW); err != nil {
			return nil, "", fmt.Errorf("grayscale: %w", err)
		}
	}

	if p.Blur > 0 {
		if err := img.GaussianBlur(p.Blur); err != nil {
			return nil, "", fmt.Errorf("blur: %w", err)
		}
	}

	if p.Effect == params.EffectSharpen {
		if err := img.Sharpen(1.0, 2.0, 0.5); err != nil {
			return nil, "", fmt.Errorf("sharpen: %w", err)
		}
	}

	format, out, err := exportImage(img, p)
	if err != nil {
		return nil, "", err
	}

	return out, format, nil
}

func applyResize(img *vips.ImageRef, p params.Params) error {
	if p.Aspect != "" {
		if err := applyAspectRatio(img, p.Aspect); err != nil {
			return err
		}
	}

	w := p.Width
	h := p.Height

	if w <= 0 && h <= 0 {
		return nil
	}

	if p.CropMode == params.CropPadResize {
		return applyPadResize(img, w, h, p)
	}

	if p.CropMode == params.CropAtMax {
		return resizeAtMax(img, w, h)
	}

	var scale float64
	if w > 0 && h > 0 {
		scaleW := float64(w) / float64(img.Width())
		scaleH := float64(h) / float64(img.Height())
		scale = math.Max(scaleW, scaleH)
		err := img.Resize(scale, vips.KernelLanczos3)
		if err != nil {
			return err
		}
		if img.Width() > w || img.Height() > h {
			cropX := (img.Width() - w) / 2
			cropY := (img.Height() - h) / 2
			if cropX < 0 {
				cropX = 0
			}
			if cropY < 0 {
				cropY = 0
			}
			return img.Crop(cropX, cropY, w, h)
		}
		return nil
	}

	if w > 0 {
		scale = float64(w) / float64(img.Width())
	} else {
		scale = float64(h) / float64(img.Height())
	}
	return img.Resize(scale, vips.KernelLanczos3)
}

func resizeAtMax(img *vips.ImageRef, w, h int) error {
	origW := img.Width()
	origH := img.Height()
	if w <= 0 {
		w = origW
	}
	if h <= 0 {
		h = origH
	}

	scaleX := float64(w) / float64(origW)
	scaleY := float64(h) / float64(origH)
	scale := math.Min(scaleX, scaleY)
	if scale >= 1 {
		return nil
	}

	return img.Resize(scale, vips.KernelLanczos3)
}

func applyPadResize(img *vips.ImageRef, w, h int, p params.Params) error {
	var scale float64
	if w > 0 && h > 0 {
		scaleW := float64(w) / float64(img.Width())
		scaleH := float64(h) / float64(img.Height())
		scale = math.Min(scaleW, scaleH)
	} else if w > 0 {
		scale = float64(w) / float64(img.Width())
		h = int(float64(img.Height()) * scale)
	} else if h > 0 {
		scale = float64(h) / float64(img.Height())
		w = int(float64(img.Width()) * scale)
	} else {
		return nil
	}

	// Escala siempre (también agranda) para que la imagen llene el marco
	// tanto como sea posible manteniendo la proporción.
	if scale != 1 {
		if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
			return err
		}
	}

	bg := parseBGColor(p.BgColor)

	// Rellena el espacio sobrante con el color de fondo, centrando la imagen.
	if img.Width() < w || img.Height() < h {
		padW := w
		padH := h
		if padW < img.Width() {
			padW = img.Width()
		}
		if padH < img.Height() {
			padH = img.Height()
		}
		left := (padW - img.Width()) / 2
		top := (padH - img.Height()) / 2
		return img.EmbedBackgroundRGBA(left, top, padW, padH, &bg)
	}

	return nil
}

func applyAspectRatio(img *vips.ImageRef, ar string) error {
	parts := strings.Split(ar, "-")
	if len(parts) != 2 {
		return nil
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if num <= 0 || den <= 0 {
		return nil
	}

	origW := float64(img.Width())
	origH := float64(img.Height())
	targetRatio := num / den
	currentRatio := origW / origH

	if currentRatio > targetRatio {
		newW := origH * targetRatio
		off := (origW - newW) / 2
		return img.Crop(int(off), 0, int(newW), img.Height())
	}

	newH := origW / targetRatio
	off := (origH - newH) / 2
	return img.Crop(0, int(off), img.Width(), int(newH))
}

func parseBGColor(hex string) vips.ColorRGBA {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 0 {
		return vips.ColorRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	c := vips.ColorRGBA{A: 255}
	switch len(hex) {
	case 6:
		fmt.Sscanf(hex, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 8:
		fmt.Sscanf(hex, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	}
	return c
}

func exportImage(img *vips.ImageRef, p params.Params) (string, []byte, error) {
	fmtParam := p.Format
	if fmtParam == "" {
		imgType := img.Format()
		switch imgType {
		case vips.ImageTypeJPEG:
			fmtParam = params.FormatJPEG
		case vips.ImageTypePNG:
			fmtParam = params.FormatPNG
		case vips.ImageTypeWEBP:
			fmtParam = params.FormatWebP
		case vips.ImageTypeAVIF:
			fmtParam = params.FormatAVIF
		case vips.ImageTypeGIF:
			fmtParam = params.FormatGIF
		default:
			fmtParam = params.FormatWebP
		}
	}

	quality := p.Quality
	if quality <= 0 {
		quality = 80
	}

	switch fmtParam {
	case params.FormatJPEG:
		ep := vips.NewJpegExportParams()
		ep.Quality = quality
		ep.StripMetadata = !p.Metadata
		if p.Progressive {
			ep.Interlace = true
		}
		out, _, err := img.ExportJpeg(ep)
		return "image/jpeg", out, err

	case params.FormatPNG:
		ep := vips.NewPngExportParams()
		ep.StripMetadata = !p.Metadata
		ep.Quality = quality
		if p.Progressive {
			ep.Interlace = true
		}
		out, _, err := img.ExportPng(ep)
		return "image/png", out, err

	case params.FormatWebP:
		ep := vips.NewWebpExportParams()
		ep.Quality = quality
		ep.StripMetadata = !p.Metadata
		if p.Lossless {
			ep.Lossless = true
		}
		out, _, err := img.ExportWebp(ep)
		return "image/webp", out, err

	case params.FormatAVIF:
		ep := vips.NewAvifExportParams()
		ep.Quality = quality
		ep.StripMetadata = !p.Metadata
		out, _, err := img.ExportAvif(ep)
		return "image/avif", out, err

	case params.FormatGIF:
		ep := vips.NewGifExportParams()
		out, _, err := img.ExportGIF(ep)
		return "image/gif", out, err

	default:
		ep := vips.NewWebpExportParams()
		ep.Quality = quality
		out, _, err := img.ExportWebp(ep)
		return "image/webp", out, err
	}
}
