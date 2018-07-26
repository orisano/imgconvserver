package imgconvserver

import (
	"context"
	"os"

	"strings"

	"io/ioutil"

	"github.com/pkg/errors"
	"sync"
)

type Fetcher interface {
	Init()
	Fetch(ctx context.Context, src *ImgSrc) ([]byte, error)
}

var fetchers = map[string]Fetcher{}

func init() {
	fetcher := &fsFetcher{}
	fetcher.Init()
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

type fsFetcher struct {
	cache *sync.Map
}

func (f *fsFetcher) Init() {
	f.cache = &sync.Map{}
}

func (f *fsFetcher) Fetch(ctx context.Context, src *ImgSrc) ([]byte, error) {
	root := src.Root
	if !strings.HasSuffix(src.Root, "/") {
		root = root + "/"
	}
	p := root + src.Path
	// if b, ok := f.imgcache.Load(p); ok {
	// 	bt, _ := b.([]byte)
	// 	buf := make([]byte, len(bt))
	// 	copy(buf, bt)
	// 	return buf, nil
	// }

	fi, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	// buf := make([]byte, len(b))
	// copy(buf, b)
	// f.imgcache.Store(p, b)

	return b, nil
}
