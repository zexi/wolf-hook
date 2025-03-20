package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zexi/wolf-hook/pkg/handlers"
	"yunion.io/x/log"
)

var (
	addr string
	port int
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1", "HTTP server listen address")
	flag.IntVar(&port, "port", 8080, "HTTP server listen port")
	flag.Parse()
}

func main() {
	log.Infof("============= WOLF HOOK ==========")
	srv := &http.Server{
		Handler:      getHandler(),
		Addr:         fmt.Sprintf("%s:%d", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Infof("Listening on %s:%d", addr, port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

func getHandler() http.Handler {
	r := mux.NewRouter()
	r.Handle("/hook/start", handlers.NewStartController())
	return r
}
