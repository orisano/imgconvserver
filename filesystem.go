package imgconvserver

import (
	"net/http"
	"os"
	"regexp"
)

type FileSystem struct {
	Conf  *DefaultConfig
	paths map[*regexp.Regexp]*Directive
}

func NewFileSystem(conf *ServerConfig) *FileSystem {
	paths := make(map[*regexp.Regexp]*Directive)
	for _, d := range conf.Directives {
		pat := regexp.MustCompile(d.UrlPattern)
		paths[pat] = &d
	}
	return &FileSystem{
		Conf:  &conf.Default,
		paths: paths,
	}
}

func (fs *FileSystem) Open(name string) (http.File, error) {
	for p, d := range fs.paths {
		matches := p.FindStringSubmatch(name)
		if len(matches) != 0 {
			continue
		}
	}

	return nil, os.ErrNotExist
}
