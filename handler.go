package imgconvserver

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"image"
	"os"

	"image/png"

	"image/jpeg"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/disintegration/imaging"
	"github.com/go-chi/chi"
	"golang.org/x/image/webp"
)

func MakeHandler(d *Directive) func(w http.ResponseWriter, r *http.Request) {
	eng, ok := engine.Engines[d.Engine]
	if !ok {
		log.Fatalf("Unsupported Engine %s", d.Function)
	}
	s, ok := eng.Specs()[d.Function]

	if !ok {
		log.Fatalf("Unsupported Function %s on Engine %s", d.Function, d.Engine)
	}
	var i interface{} = eng

	return func(w http.ResponseWriter, r *http.Request) {
		filepath := chi.URLParam(r, ":filepath")
		im, err := openImage(filepath)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		switch d.Function {
		case "Resize":
			i2 := i.(engine.Resizer)
			x := chi.URLParam(r, ":dx")
			y := chi.URLParam(r, ":dy")

			dx, err := strconv.Atoi(x)
			if err != nil {
				http.Error(w, http.StatusText(400), 400)
				return
			}
			dy, err := strconv.Atoi(y)
			if err != nil {
				http.Error(w, http.StatusText(400), 400)
				return
			}
			im, err = i2.Resize(im, dx, dy)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
		of := chi.URLParam(r, ":format")
		switch of {
		case "png":
			if err := png.Encode(w, im); err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		case "jpeg", "JPEG", "jpg":
			if err := jpeg.Encode(w, im, nil); err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		case "webp":
			if err := webp.Encode(w, im, nil); err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
	}
}

func openImage(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	i, _, err := image.Decode(f)
	return i, err
}

func ResizeHandler(w http.ResponseWriter, r *http.Request) {
	dx := chi.URLParam(r, "dx")
	dy := chi.URLParam(r, "dy")
	imagename := chi.URLParam(r, "imagename")

	opt := DefaultHandlerOpt()
	x, err := strconv.Atoi(dx)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	y, err := strconv.Atoi(dy)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if err := resize(w, opt.Mount+imagename, x, y); err != nil {
		http.Error(w, http.StatusText(404), 404)
		log.Printf("convert error %+v", err)
		return
	}
}

func resize(w io.Writer, imagepath string, dx, dy int) error {
	src, err := imaging.Open(imagepath)
	if err != nil {
		return err
	}
	dist := imaging.Resize(src, dx, dy, imaging.Lanczos)
	return imaging.Encode(w, dist, imaging.JPEG)
}
