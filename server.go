package imgconvserver

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
	"golang.org/x/sync/singleflight"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/akito0107/imgconvserver/format"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

var cache sync.Map
var imgcache sync.Map
var sfg singleflight.Group
var isfg singleflight.Group

type handler struct {
	conf  *DefaultConfig
	paths map[*regexp.Regexp]Directive

	semaphore chan struct{}
}

type record struct {
	storedAt time.Time
	buf      []byte
}

func Server(conf *ServerConfig) http.Handler {
	paths := make(map[*regexp.Regexp]Directive)
	for _, d := range conf.Directives {
		pat := regexp.MustCompile(d.UrlPattern)
		paths[pat] = d
	}
	return &handler{
		semaphore: make(chan struct{}, 32),

		conf:  &conf.Default,
		paths: paths,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	log.Println(upath)

	for p, d := range h.paths {
		matches := p.FindStringSubmatch(upath)
		if len(matches) == 0 {
			continue
		}

		vars := map[string]interface{}{}
		for i, m := range matches {
			s := fmt.Sprintf("$%d", i)
			vars[s] = m
		}

		for k, v := range d.Vars {
			str, ok := v.(string)
			if !ok {
				vars[k] = v
			}
			if strings.HasPrefix(str, "$") {
				val, ok := vars[str]
				if ok {
					vars[k] = val
				}
			}
		}
		ctx := context.WithValue(r.Context(), "vars", vars)
		ctx = context.WithValue(ctx, "drc", d)
		ctx = context.WithValue(ctx, "upath", upath)

		t := time.NewTimer(3 * time.Second)
		select {
		case h.semaphore <- struct{}{}:
			serve(w, r.WithContext(ctx))
			<-h.semaphore
			t.Stop()
		case <-t.C:
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			t.Stop()
		}
		return
	}
	http.Error(w, http.StatusText(404), 404)
}

func serve(w http.ResponseWriter, r *http.Request) {
	drc := r.Context().Value("drc").(Directive)
	vars := r.Context().Value("vars").(map[string]interface{})
	upath := r.Context().Value("upath").(string)

	var rec *record
	if res, ok := cache.Load(upath); ok {
		rec = res.(*record)
	} else {
		res, err, _ := sfg.Do(upath, func() (interface{}, error) {
			src := &ImgSrc{
				Type: getOptValueString(drc.Src.Type, vars),
				Root: getOptValueString(drc.Src.Root, vars),
				Path: getOptValueString(drc.Src.Path, vars),
				Url:  getOptValueString(drc.Src.Url, vars),
			}

			var im image.Image
			if res, ok := imgcache.Load(src.Path); ok {
				im = res.(image.Image)
			} else {
				res, err, _ := isfg.Do(src.Path, func() (interface{}, error) {
					file, err := Fetch(r.Context(), src)
					if err != nil {
						return nil, &convError{
							code: http.StatusNotFound,
							err:  err,
						}
					}
					im, err := decodeImage(upath, file)
					if err != nil {
						return nil, &convError{
							code: http.StatusBadRequest,
							err:  err,
						}
					}
					imgcache.Store(src.Path, im)
					return im, nil
				})
				if err != nil {
					return nil, err
				}
				im = res.(image.Image)
			}

			engtype := getOptValueString(drc.Engine, vars)
			eng, ok := engine.Engines[engtype]
			if !ok {
				log.Printf("Unsupported Engine %s", engtype)
				return nil, &convError{
					code: http.StatusBadRequest,
				}
			}

			opt := &engine.ConvertOptions{}
			if err := completionOptions(opt, drc, vars); err != nil {
				return nil, &convError{
					code: http.StatusBadRequest,
					err:  err,
				}
			}
			resizer, ok := eng.(engine.Resizer)
			if !ok && opt.Resize {
				return nil, &convError{
					code: http.StatusBadRequest,
				}
			}
			encoder, ok := eng.(engine.Encoder)
			if !ok {
				log.Println("using Default Encoder")
				encoder = &engine.DefaultEncoder{}
			}

			var err error
			for _, c := range drc.Converts {
				switch c.Type {
				case "resize":
					im, err = resizer.Resize(im, &engine.ResizeOptions{
						Dw: opt.Dw,
						Dh: opt.Dh,
					})
					if err != nil {
						return nil, &convError{
							code: http.StatusInternalServerError,
							err:  err,
						}
					}
				default:
					log.Printf("unsupported function pattern %s\n", c.Type)
					return nil, &convError{
						code: http.StatusBadRequest,
					}
				}
			}

			var b bytes.Buffer
			if err := encoder.Encode(&b, im, &engine.EncodeOptions{
				Format:  opt.Format,
				Quality: opt.Quality,
			}); err != nil {
				return nil, &convError{
					code: http.StatusBadRequest,
					err:  err,
				}
			}
			rec := &record{buf: b.Bytes(), storedAt: time.Now()}
			cache.Store(upath, rec)
			return rec, nil
		})
		if err != nil {
			err.(*convError).Write(w)
			return
		}
		rec = res.(*record)
	}
	w.Header().Set("Last-Modified", rec.storedAt.Format(TimeFormat))
	if _, err := w.Write(rec.buf); err != nil {
		log.Printf("write error: %+v", err)
	}
}

func getOptValue(value interface{}, vars map[string]interface{}) interface{} {
	v, ok := value.(string)
	if ok && strings.HasPrefix(v, "$") {
		return vars[v]
	}
	return value
}

func getOptValueString(value string, vars map[string]interface{}) string {
	if strings.HasPrefix(value, "$") {
		return vars[value].(string)
	}
	return value
}

func completionOptions(opt *engine.ConvertOptions, drc Directive, vars map[string]interface{}) (err error) {
	opt.Format = format.FromString(drc.Output.Format)
	opt.Quality = drc.Output.Quality

	for _, c := range drc.Converts {
		switch c.Type {
		case "resize":
			opt.Resize = true
			wid := getOptValue(c.Parameters["dw"], vars)
			hgt := getOptValue(c.Parameters["dh"], vars)

			dw, ok := wid.(int)
			if !ok {
				wd := wid.(string)
				dw, err = strconv.Atoi(wd)
				if err != nil {
					return
				}
			}

			dh, ok := hgt.(int)
			if !ok {
				ht := hgt.(string)
				dh, err = strconv.Atoi(ht)
				if err != nil {
					return
				}
			}
			opt.Dh = dh
			opt.Dw = dw
		}
	}

	return nil
}

func decodeImage(path string, file []byte) (image.Image, error) {
	f := bytes.NewReader(file)
	im, _, err := image.Decode(f)
	return im, err
}

type convError struct {
	code int
	err  error
}

func (ce *convError) Error() string {
	if ce.err != nil {
		return ce.err.Error()
	}
	return http.StatusText(ce.code)
}

func (ce *convError) Write(w http.ResponseWriter) {
	http.Error(w, ce.Error(), ce.code)
}
