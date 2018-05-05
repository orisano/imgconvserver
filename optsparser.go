package imgconvserver

import (
	"io"
	"log"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Default    DefaultConfig `toml:"default"`
	Directives []Directive   `toml:"directives"`
}

type DefaultConfig struct {
	Src *ImgSrc `toml:"src"`
}

type ImgSrc struct {
	Type string `toml:"type"`
	Root string `toml:"root"`
	Path string `toml:"path"`
	Url  string `toml:"host"`
}

type Directive struct {
	Name       string    `toml:"name"`
	Engine     string    `toml:"engine"`
	UrlPattern string    `toml:"urlpattern"`
	Input      string    `toml:"input"`
	Src        ImgSrc    `toml:"src"`
	Format     string    `toml:"string"`
	Converts   []Convert `toml:"converts"`
	Vars       map[string]interface{}
}

type Convert struct {
	Type       string                 `toml:"type"`
	Parameters map[string]interface{} `toml:"parameters"`
}

func Parse(r io.Reader) *ServerConfig {
	var conf ServerConfig
	_, err := toml.DecodeReader(r, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return &conf
}
