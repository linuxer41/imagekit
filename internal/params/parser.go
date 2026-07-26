package params

import (
	"strconv"
	"strings"
)

func ParseTR(raw string) Params {
	p := Params{}
	if raw == "" {
		return p
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		idx := strings.IndexByte(part, '-')
		if idx < 1 {
			continue
		}
		key := part[:idx]
		val := part[idx+1:]

		switch key {
		case "w":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Width = v
				p.hasParams = true
			}
		case "h":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Height = v
				p.hasParams = true
			}
		case "ar":
			p.Aspect = val
			p.hasParams = true
		case "cm":
			p.CropMode = CropMode(val)
			p.hasParams = true
		case "q":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Quality = v
				p.hasParams = true
			}
		case "f":
			p.Format = Format(val)
			p.hasParams = true
		case "bg":
			p.BgColor = val
			p.hasParams = true
		case "rt":
			if v, err := strconv.Atoi(val); err == nil {
				p.Rotation = v % 360
				p.hasParams = true
			}
		case "fo":
			p.Focus = Focus(val)
			p.hasParams = true
		case "bl":
			if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
				p.Blur = v
				p.hasParams = true
			}
		case "e":
			p.Effect = Effect(val)
			p.hasParams = true
		case "oi":
			p.OverlayImage = val
			p.hasParams = true
		case "ot":
			p.OverlayText = val
			p.hasParams = true
		case "ox":
			if v, err := strconv.Atoi(val); err == nil {
				p.OverlayX = v
				p.hasParams = true
			}
		case "oy":
			if v, err := strconv.Atoi(val); err == nil {
				p.OverlayY = v
				p.hasParams = true
			}
		case "ow":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.OverlayW = v
				p.hasParams = true
			}
		case "oh":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.OverlayH = v
				p.hasParams = true
			}
		case "b":
			p.Border = val
			p.hasParams = true
		case "pr":
			p.Progressive = val == "true"
			p.hasParams = true
		case "lo":
			p.Lossless = val == "true"
			p.hasParams = true
		case "md":
			p.Metadata = val == "true"
			p.hasParams = true
		case "t":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Trim = v
				p.hasParams = true
			}
		case "l":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Radius = v
				p.hasParams = true
			}
		case "pg":
			if v, err := strconv.Atoi(val); err == nil && v > 0 {
				p.Page = v
				p.hasParams = true
			}
		}
	}

	return p
}
