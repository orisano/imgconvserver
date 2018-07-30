package imgconvserver

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

type Fetcher interface {
	Fetch(ctx context.Context, src *ImgSrc) ([]byte, error)
}

var fetchers = map[string]Fetcher{}

func init() {
	fetcher := &fsFetcher{}
	fetchers["fs"] = fetcher
}

func Fetch(ctx context.Context, src *ImgSrc) ([]byte, error) {
	typ := src.Type
	fetcher, ok := fetchers[typ]
	if !ok {
		return nil, errors.Errorf("fetcher %s not found", typ)
	}
	return fetcher.Fetch(ctx, src)
}

type fsFetcher struct{}

func (f *fsFetcher) Fetch(ctx context.Context, src *ImgSrc) ([]byte, error) {
	root := src.Root
	if !strings.HasSuffix(src.Root, "/") {
		root = root + "/"
	}
	p := root + src.Path
	return ioutil.ReadFile(p)
}
