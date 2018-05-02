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

func (engine) Specs() map[string]interface{} {
	return specs
}

func (engine) Resize(src image.Image, dx, dy int) (image.Image, error) {
	dist := imaging.Resize(src, dx, dy, imaging.Lanczos)

	return dist, nil
}
