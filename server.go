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

	"strconv"

	"github.com/akito0107/imgconvserver/engine"
	"github.com/akito0107/imgconvserver/format"
	_ "golang.org/x/image/webp"
)

type handler struct {
	conf  *DefaultConfig
	paths map[*regexp.Regexp]Directive
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
			ctx = context.WithValue(ctx, "opt", d)

			serve(w, r.WithContext(ctx))
			return
		}
	}
	http.Error(w, http.StatusText(404), 404)
}

func serve(w http.ResponseWriter, r *http.Request) {
	opt := r.Context().Value("opt").(Directive)
	vars := r.Context().Value("vars").(map[string]interface{})
	src := &ImgSrc{
		Type: getOptValueString(opt.Src.Type, vars),
		Root: getOptValueString(opt.Src.Root, vars),
		Path: getOptValueString(opt.Src.Path, vars),
		Url:  getOptValueString(opt.Src.Url, vars),
	}

	file, err := Fetch(r.Context(), src)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), 404)
		return
	}

	im, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	engtype := getOptValueString(opt.Engine, vars)
	eng, ok := engine.Engines[engtype]
	if !ok {
		log.Printf("Unsupported Engine %s", engtype)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	resizer, ok := eng.(engine.Resizer)
	for _, c := range opt.Converts {
		if !ok && c.Type == "resize" {
			log.Printf("Unsupported Resize function on engine %s", engtype)
			http.Error(w, http.StatusText(400), 400)
			return
		}
	}

	encoder, ok := eng.(engine.Encoder)
	if !ok {
		log.Println("using Default Encoder")
		encoder = &engine.DefaultEncoder{}
	}

	for _, c := range opt.Converts {
		switch c.Type {
		case "resize":
			wid := getOptValue(c.Parameters["dw"], vars)
			hgt := getOptValue(c.Parameters["dh"], vars)

			dw, ok := wid.(int)
			if !ok {
				wd := wid.(string)
				dw, err = strconv.Atoi(wd)
				if err != nil {
					http.Error(w, http.StatusText(400), 400)
					return
				}
			}

			dh, ok := hgt.(int)
			if !ok {
				ht := hgt.(string)
				dh, err = strconv.Atoi(ht)
				if err != nil {
					http.Error(w, http.StatusText(400), 400)
					return
				}
			}

			im, err = resizer.Resize(im, dw, dh)
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
	of := opt.Format
	op := &engine.EncodeOptions{
		Format: format.FromString(of),
	}
	encoder.Encode(w, im, op)
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
