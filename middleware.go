package imgconvserver

import (
	"image"
	"log"
	"net/http"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"strings"

	"net/url"

	"github.com/go-chi/chi"
	_ "golang.org/x/image/webp"
)

func OptionParser(def *DefaultConfig, d *Directive) func(http.Handler) http.Handler {
	src := d.Src
	if def.Src == nil && src == nil {
		log.Fatal("no image src provided")
	}
	if src == nil {
		src = def.Src
	}

	url.ParseRequestURI(d.UrlPattern)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
	}
}

func isVariableParam(in string) bool {
	return strings.HasPrefix(in, ":")
}

func getQueryOrURLParam(r *http.Request, key string) string {
	param := chi.URLParam(r, key)
	if param == "" {
	}
}

func openImage(src *ImgSrc, filepath string) (image.Image, error) {
	f, err := os.Open(src.Root + filepath)
	if err != nil {
		return nil, err
	}
	i, _, err := image.Decode(f)
	return i, err
}
