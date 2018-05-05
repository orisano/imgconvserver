package engine

import (
	"image"
	"io"

	"image/jpeg"
	"image/png"

	"github.com/akito0107/imgconvserver/format"
	"github.com/pkg/errors"
)

var Engines = map[string]interface{}{}

func Register(name string, engine interface{}) {
	Engines[name] = engine
}

type Engine interface {
	Convert(src image.Image, opts *ConvertOptions) (image.Image, error)
}

type Converter interface {
	Convert(src image.Image, opts *ConvertOptions) (image.Image, error)
}

type Resizer interface {
	Resize(src image.Image, dw, dh int) (image.Image, error)
}

type Encoder interface {
	Encode(w io.Writer, src image.Image, opt *EncodeOptions) error
}

type DefaultEncoder struct{}

func (d *DefaultEncoder) Encode(w io.Writer, src image.Image, opt *EncodeOptions) error {
	switch opt.Format {
	case format.PNG:
		if err := png.Encode(w, src); err != nil {
			return errors.Errorf("encoding error %+v", err)
		}
	case format.JPEG:
		if err := jpeg.Encode(w, src, &jpeg.Options{
			Quality: opt.Quality,
		}); err != nil {
			return errors.Errorf("encoding error %+v", err)
		}
	default:
		return errors.Errorf("unsupported encode format")
	}
	return nil
}

type ConvertOptions struct {
	ResizeOptions
	EncodeOptions
}

type ResizeOptions struct {
	Resize bool
	Dh     int
	Dw     int
}

type EncodeOptions struct {
	Format  format.Format
	Quality int
}

type ConvertOption func(options *ConvertOptions) error

func Dw(w int) ConvertOption {
	return func(opts *ConvertOptions) error {
		opts.Dw = w
		return nil
	}
}

func Dh(h int) ConvertOption {
	return func(opts *ConvertOptions) error {
		opts.Dh = h
		return nil
	}
}
