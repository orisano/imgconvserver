package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/akito0107/imgconvserver"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	_ "github.com/akito0107/imgconvserver/engine/imaging"
)

var configpath = flag.String("conf", "conf.toml", "config file path (default: conf.toml)")

func main() {
	flag.Parse()

	f, err := os.Open(*configpath)
	if err != nil {
		log.Fatalf("config file open failed %+v", err)
	}
	conf := imgconvserver.Parse(f)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	for _, d := range conf.Directives {
		r.Get(d.UrlPattern, imgconvserver.MakeHandler(&d))
	}

	http.ListenAndServe(":3000", r)
}
