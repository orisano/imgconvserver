package engine

import "image"

var Engines = map[string]Engine{}

func Register(name string, engine Engine) {
	Engines[name] = engine
}

type Engine interface {
	Specs() map[string]interface{}
}

type Resizer interface {
	Resize(src image.Image, dx, dy int) (image.Image, error)
}
