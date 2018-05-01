package imaging

import (
	"image"

	eng "github.com/akito0107/imgconvserver/engine"
	"github.com/disintegration/imaging"
)

func init() {
	spec := &eng.Spec{
		SupportedFuncs: []string{"resize"},
	}
	eng.Register("imaging", &engine{}, spec)
}

type engine struct{}

func (e *engine) Resize(src image.Image, dx, dy int) (image.Image, error) {
	dist := imaging.Resize(src, dx, dy, imaging.Lanczos)

	return dist, nil
}
