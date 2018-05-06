package imaging

import (
	"image"

	eng "github.com/akito0107/imgconvserver/engine"
	"github.com/disintegration/imaging"
)

func init() {
	eng.Register("imaging", &engine{})
}

type engine struct{}

func (engine) Resize(src image.Image, dw, dh int) (image.Image, error) {
	dist := imaging.Resize(src, dw, dh, imaging.Lanczos)
	return dist, nil
}
