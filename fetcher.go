package imgconvserver

import (
	"context"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

type Fetcher interface {
	Fetch(ctx context.Context, src *ImgSrc) (http.File, error)
}

var fetchers = map[string]Fetcher{}

func init() {
	fetchers["fs"] = &fsFetcher{}
}

func Fetch(ctx context.Context, src *ImgSrc) (http.File, error) {
	typ := src.Type
	fetcher, ok := fetchers[typ]
	if !ok {
		return nil, errors.Errorf("fetcher %s not found", typ)
	}
	return fetcher.Fetch(ctx, src)
}

type fsFetcher struct{}

func (fsFetcher) Fetch(ctx context.Context, src *ImgSrc) (http.File, error) {
	return os.Open(src.Root + src.Path)
}
