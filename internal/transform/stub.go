//go:build !cgo

package transform

import (
	"fmt"

	"github.com/vendemas/imagekit/internal/params"
)

func ProcessImage(input []byte, p params.Params) ([]byte, string, error) {
	return nil, "", fmt.Errorf("image processing requires CGO (libvips)")
}
