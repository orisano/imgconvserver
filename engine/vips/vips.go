package vips

import (
	"image"

	"bytes"

	"image/jpeg"

	eng "github.com/akito0107/imgconvserver/engine"
	"github.com/daddye/vips"
	"io"
)

type engine struct{}

func (engine) Convert(w io.Writer, src []byte, opts *eng.ConvertOptions) error {
	opt := vips.Options{
		Width:        opts.Dw,
		Height:       opts.Dh,
		Crop:         false,
		Extend:       vips.EXTEND_WHITE,
		Interpolator: vips.BILINEAR,
		Gravity:      vips.CENTRE,
		Quality:      95,
	}
	out, err := vips.Resize(src, opt)
	if err != nil {
		return err
	}
	_, err = w.Write(out)

	return err
}

func init() {
	eng.Register("vips", &engine{})
}

func (engine) Resize(src image.Image, dw, dh int) (image.Image, error) {
	var buf bytes.Buffer
	opts := vips.Options{
		Width:        dw,
		Height:       dh,
		Crop:         false,
		Extend:       vips.EXTEND_WHITE,
		Interpolator: vips.BILINEAR,
		Gravity:      vips.CENTRE,
		Quality:      95,
	}
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 100}); err != nil {
		return nil, err
	}
	out, err := vips.Resize(buf.Bytes(), opts)
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(bytes.NewReader(out))
}
