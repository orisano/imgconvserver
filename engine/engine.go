package engine

import "image"

var Engines = map[string]Engine{}
var Specs = map[string]*Spec{}

func Register(name string, engine Engine, spec *Spec) {
	Engines[name] = engine
	Specs[name] = spec
}

type Spec struct {
	SupportedFuncs []string
}

type ResizeOpts struct {
}

type ResizeOpt func(*ResizeOpts)

type Engine interface {
	Resize(src image.Image, dx, dy int) (image.Image, error)
}
