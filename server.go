package imgconvserver

import (
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"bytes"
	"strconv"
	"time"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/akito0107/imgconvserver/format"
	"sync"
	"io"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

var cache sync.Map
var imgcache sync.Map

type handler struct {
	conf  *DefaultConfig
	paths map[*regexp.Regexp]Directive
}

type record struct {
	storedAt time.Time
	buf      bytes.Buffer
}

func Server(conf *ServerConfig) http.Handler {
	paths := make(map[*regexp.Regexp]Directive)
	for _, d := range conf.Directives {
		pat := regexp.MustCompile(d.UrlPattern)
		paths[pat] = d
	}
	return &handler{
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

	rec, ok := cache.Load(upath)
	if ok {
		record, ok := rec.(*record)
		if !ok {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		w.Header().Set("Last-Modified", record.storedAt.Format(TimeFormat))
		b := make([]byte, record.buf.Len())
		copy(b, record.buf.Bytes())
		io.Copy(w, bytes.NewBuffer(b))
		return
	}

	for p, d := range h.paths {
		matches := p.FindStringSubmatch(upath)
		if len(matches) != 0 {
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

			serve(w, r.WithContext(ctx))
			return
		}
	}
	http.Error(w, http.StatusText(404), 404)
}

func serve(w http.ResponseWriter, r *http.Request) {
	drc := r.Context().Value("drc").(Directive)
	vars := r.Context().Value("vars").(map[string]interface{})
	upath := r.Context().Value("upath").(string)

	src := &ImgSrc{
		Type: getOptValueString(drc.Src.Type, vars),
		Root: getOptValueString(drc.Src.Root, vars),
		Path: getOptValueString(drc.Src.Path, vars),
		Url:  getOptValueString(drc.Src.Url, vars),
	}

	file, err := Fetch(r.Context(), src)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}

	engtype := getOptValueString(drc.Engine, vars)
	eng, ok := engine.Engines[engtype]
	if !ok {
		log.Printf("Unsupported Engine %s", engtype)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	opt := &engine.ConvertOptions{}

	if err := completionOptions(opt, drc, vars); err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	// if cvt, ok := eng.(engine.Converter); ok {
	// 	if err := cvt.Convert(w, file, opt); err != nil {
	// 		http.Error(w, http.StatusText(500), 500)
	// 	}
	// 	return
	// }

	resizer, ok := eng.(engine.Resizer)
	if !ok && opt.Resize {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	encoder, ok := eng.(engine.Encoder)
	if !ok {
		log.Println("using Default Encoder")
		encoder = &engine.DefaultEncoder{}
	}

	im, err := decodeImage(upath, file)

	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}
	for _, c := range drc.Converts {
		switch c.Type {
		case "resize":
			im, err = resizer.Resize(im, &engine.ResizeOptions{
				Dw: opt.Dw,
				Dh: opt.Dh,
			})
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		default:
			log.Printf("unsupported function pattern %s\n", c.Type)
			http.Error(w, http.StatusText(400), 400)
			return
		}
	}
	now := time.Now()
	w.Header().Set("Last-Modified", now.Format(TimeFormat))
	var buf bytes.Buffer
	wr := io.MultiWriter(&buf, w)
	encoder.Encode(wr, im, &engine.EncodeOptions{
		Format:  opt.Format,
		Quality: opt.Quality,
	})

	cache.Store(upath, &record{storedAt: now, buf: buf})
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
	if im, ok := imgcache.Load(path); ok {
		i := im.(image.Image)
		return i, nil
	}
	f := bytes.NewBuffer(file)
	im, _, err := image.Decode(f)
	imgcache.Store(path, im)
	return im, err
}
