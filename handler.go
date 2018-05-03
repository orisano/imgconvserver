package imgconvserver

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"image"
	"os"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/disintegration/imaging"
	"github.com/go-chi/chi"
)

func MakeHandler(d *Directive) func(w http.ResponseWriter, r *http.Request) {
	eng, ok := engine.Engines[d.Engine]
	if !ok {
		log.Fatalf("Unsupported Engine %s", d.Function)
	}
	resizer, ok := eng.(engine.Resizer)
	if !ok && d.Function == "resize" {
		log.Fatalf("Unsupported Resize function on engine %s", d.Engine)
	}

	encoder, ok := eng.(engine.Encoder)
	if !ok {
		log.Println("using Default Encoder")
		encoder = &engine.DefaultEncoder{}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		filepath := chi.URLParam(r, ":filepath")
		im, err := openImage(filepath)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		switch d.Function {
		case "Resize":
			wid := chi.URLParam(r, ":dw")
			hgt := chi.URLParam(r, ":dh")

			dw, err := strconv.Atoi(wid)
			if err != nil {
				http.Error(w, http.StatusText(400), 400)
				return
			}
			dh, err := strconv.Atoi(hgt)
			if err != nil {
				http.Error(w, http.StatusText(400), 400)
				return
			}
			im, err = resizer.Resize(im, dw, dh)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
		of := d.Format
		op := &engine.EncodeOptions{
			Format: FromString(of),
		}
		encoder.Encode(w, im, op)
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
