package imgconvserver

import (
	"io"
	"log"

	"github.com/BurntSushi/toml"
)

type ServerConfig struct {
	Mount      string      `toml:"mount"`
	Directives []Directive `toml:"directives"`
}

type Directive struct {
	UrlPattern string                 `toml:"urlpattern"`
	Format     string                 `toml:"format"`
	Function   string                 `toml:"function"`
	Parameters map[string]interface{} `toml:"parameters"`
}

func DefaultHandlerOpt() *ServerConfig {
	return &ServerConfig{
		Mount: "./",
	}
}

func Parse(r io.Reader) *ServerConfig {
	var conf ServerConfig
	_, err := toml.DecodeReader(r, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return &conf
}
