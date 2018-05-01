package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"time"
	"github.com/akito0107/imgconvserver"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Get("/resize/{dx}/{dy}/{imagename}", imgconvserver.ResizeHandler)

	http.ListenAndServe(":3000", r)
}
