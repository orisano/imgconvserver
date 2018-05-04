package imgconvserver

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

	for p, d := range h.paths {
		matches := p.FindStringSubmatch(upath)
		if len(matches) != 0 {
			vals := map[string]interface{}{}
			for i, m := range matches {
				s := fmt.Sprintf("$%d", i)
				vals[s] = m
			}

			for k, v := range d.Values {
				str, ok := v.(string)
				if !ok {
					vals[k] = v
				}
				if strings.HasPrefix(str, "$") {
					val, ok := vals[str]
					if ok {
						vals[k] = val
					}
				}
			}
			serveFile(w, r, vals, d)
			return
		}
	}
	http.Error(w, http.StatusText(404), 404)
}

func serveFile(w http.ResponseWriter, r *http.Request, vals map[string]interface{}, opt Directive) {
	file, err := Fetch(&opt.Src)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
}

func completionOpts(vars map[string]string, opt *Directive) error {
}
