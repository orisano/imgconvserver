package imgconvserver

import (
	"context"
	"os"

	"strings"

	"github.com/pkg/errors"
	"io/ioutil"
)

type Fetcher interface {
	Fetch(ctx context.Context, src *ImgSrc) ([]byte, error)
}

var fetchers = map[string]Fetcher{}

func init() {
	fetchers["fs"] = &fsFetcher{}
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

func (fsFetcher) Fetch(ctx context.Context, src *ImgSrc) ([]byte, error) {
	root := src.Root
	if !strings.HasSuffix(src.Root, "/") {
		root = root + "/"
	}
	f, err :=  os.Open(root + src.Path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}
