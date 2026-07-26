package transform

import "fmt"

type CropMode string

const (
	CropForce           CropMode = "force"
	CropAtMax           CropMode = "at_max"
	CropAtLeast         CropMode = "at_least"
	CropMaxSizeEnum     CropMode = "max_size_enum"
	CropMaintainRatio   CropMode = "maintain_ratio"
	CropPadResize       CropMode = "pad_resize"
	CropExtract         CropMode = "extract"
	CropCrop            CropMode = "crop"
	CropTrim            CropMode = "trim"
)

type Focus string

const (
	FocusAuto           Focus = "auto"
	FocusCenter         Focus = "center"
	FocusTop            Focus = "top"
	FocusLeft           Focus = "left"
	FocusRight          Focus = "right"
	FocusBottom         Focus = "bottom"
	FocusTopLeft        Focus = "top_left"
	FocusTopRight       Focus = "top_right"
	FocusBottomLeft     Focus = "bottom_left"
	FocusBottomRight    Focus = "bottom_right"
	FocusFace           Focus = "face"
)

type Effect string

const (
	EffectSharpen   Effect = "sharpen"
	EffectGrayscale Effect = "grayscale"
	EffectContrast  Effect = "contrast"
	EffectBright    Effect = "bright"
)

type Format string

const (
	FormatJPEG Format = "jpg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatAVIF Format = "avif"
	FormatGIF  Format = "gif"
)

type Params struct {
	Width     int     `json:"w,omitempty"`
	Height    int     `json:"h,omitempty"`
	Aspect    string  `json:"ar,omitempty"`
	CropMode  CropMode `json:"cm,omitempty"`
	Quality   int     `json:"q,omitempty"`
	Format    Format  `json:"f,omitempty"`
	BgColor   string  `json:"bg,omitempty"`
	Rotation  int     `json:"rt,omitempty"`
	Focus     Focus   `json:"fo,omitempty"`
	Blur      float64 `json:"bl,omitempty"`
	Effect    Effect  `json:"e,omitempty"`
	OverlayImage string `json:"oi,omitempty"`
	OverlayText  string `json:"ot,omitempty"`
	OverlayX  int     `json:"ox,omitempty"`
	OverlayY  int     `json:"oy,omitempty"`
	OverlayW  int     `json:"ow,omitempty"`
	OverlayH  int     `json:"oh,omitempty"`
	Border    string  `json:"b,omitempty"`
	Progressive bool  `json:"pr,omitempty"`
	Lossless    bool  `json:"lo,omitempty"`
	Metadata    bool  `json:"md,omitempty"`
	Trim       int     `json:"t,omitempty"`
	Radius     int     `json:"l,omitempty"`
	Page       int     `json:"pg,omitempty"`

	hasParams bool
}

func (p Params) HasParams() bool {
	return p.hasParams
}

func (p Params) CacheKey() string {
	return fmt.Sprintf("%d|%d|%s|%s|%d|%s|%s|%d|%s|%f|%s|%s|%s|%d|%d|%d|%d|%s|%t|%t|%t|%d|%d|%d",
		p.Width, p.Height, p.Aspect, string(p.CropMode), p.Quality,
		string(p.Format), p.BgColor, p.Rotation, string(p.Focus),
		p.Blur, string(p.Effect), p.OverlayImage, p.OverlayText,
		p.OverlayX, p.OverlayY, p.OverlayW, p.OverlayH,
		p.Border, p.Progressive, p.Lossless, p.Metadata,
		p.Trim, p.Radius, p.Page)
}
