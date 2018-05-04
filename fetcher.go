package imgconvserver

import (
	"log"
	"net/http"
	"os"
)

type Fetcher interface {
	Fetch(src *ImgSrc) (http.File, error)
}

var fetchers = map[string]Fetcher{}

func init() {
	fetchers["fs"] = &fsFetcher{}
}

func Fetch(src *ImgSrc) (http.File, error) {
	typ := src.Type
	fetcher, ok := fetchers[typ]
	if !ok {
		log.Fatal("Fetcher Not found")
	}
	return fetcher.Fetch(src)
}

type fsFetcher struct{}

func (fsFetcher) Fetch(src *ImgSrc) (http.File, error) {
	return os.Open(src.Root + src.Path)
}
