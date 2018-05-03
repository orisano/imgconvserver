package imaging

import (
	"image"

	eng "github.com/akito0107/imgconvserver/engine"
	"github.com/disintegration/imaging"
)

var specs = make(map[string]interface{})

func init() {
	eng.Register("imaging", &engine{})
	specs["Resize"] = map[string]interface{}{
		"dx": "int",
		"dy": "int",
	}
}

type engine struct{}

func (engine) Convert(src image.Image, opts *eng.ConvertOptions) (image.Image, error) {
	if opts.Resize {
		src = imaging.Resize(src, opts.Dw, opts.Dh, imaging.Lanczos)
	}
	return src, nil
}

func (engine) Specs() map[string]interface{} {
	return specs
}

func (engine) Resize(src image.Image, dw, dh int) (image.Image, error) {
	dist := imaging.Resize(src, dw, dh, imaging.Lanczos)
	return dist, nil
}
