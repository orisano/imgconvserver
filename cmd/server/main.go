package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/akito0107/imgconvserver"
	_ "github.com/akito0107/imgconvserver/engine/imaging"
	_ "github.com/akito0107/imgconvserver/engine/vips"

	_ "net/http/pprof"
)

var configpath = flag.String("conf", "conf.toml", "config file path (default: conf.toml)")

func main() {
	flag.Parse()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	f, err := os.Open(*configpath)
	if err != nil {
		log.Fatalf("config file open failed %+v", err)
	}

	conf := imgconvserver.Parse(f)
	log.Fatal(http.ListenAndServe(":8080", imgconvserver.Server(conf)))
}
