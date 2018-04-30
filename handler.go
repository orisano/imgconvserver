package imgconvserver

import (
	"net/http"
	"github.com/go-chi/chi"
	"io"
	"log"
	"strconv"
	"github.com/disintegration/imaging"
)

type HandlerOpt struct {
	Base string
}

func DefaultHandlerOpt() *HandlerOpt {
	return &HandlerOpt{
		Base: "./",
	}
}


func ResizeHandler(w http.ResponseWriter, r *http.Request)  {
	dx := chi.URLParam(r, "dx")
	dy := chi.URLParam(r, "dy")
	imagename := chi.URLParam(r, "imagename")

	opt := DefaultHandlerOpt()
	x, err := strconv.ParseInt(dx, 0,64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	y, err := strconv.ParseInt(dy, 0,64)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if err := resize(w, opt.Base + imagename, x, y); err != nil {
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