package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"fmt"
		_ "net/http/pprof"

	"github.com/akito0107/imgconvserver"
	_ "github.com/akito0107/imgconvserver/engine/imaging"
	"time"
	"runtime/debug"
)

var configpath = flag.String("conf", "conf.toml", "config file path (default: conf.toml)")
var port = flag.Int("port", 8080, "listen port")
var quiet = flag.Bool("q", false, "quiet")

func main() {
	flag.Parse()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	go func() {
		t := time.NewTicker(10 * time.Second)
		for range t.C {
			debug.FreeOSMemory()
		}
	}()
	if *quiet {
		log.SetOutput(ioutil.Discard)
	}

	f, err := os.Open(*configpath)
	if err != nil {
		log.Fatalf("config file open failed %+v", err)
	}

	conf := imgconvserver.Parse(f)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), imgconvserver.Server(conf)))
}
