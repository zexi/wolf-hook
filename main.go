package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"yunion.io/x/pkg/errors"

	"github.com/zexi/wolf-hook/pkg/handlers"
	"github.com/zexi/wolf-hook/pkg/util/procutils"

	"yunion.io/x/log"
)

var (
	addr             string
	port             int
	ulimitNofileHard int
	ulimitNofileSoft int
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1", "HTTP server listen address")
	flag.IntVar(&port, "port", 8080, "HTTP server listen port")
	flag.IntVar(&ulimitNofileHard, "ulimit-nofile-hard", 10240, "ulimit nofile hard")
	flag.IntVar(&ulimitNofileSoft, "ulimit-nofile-soft", 10240, "ulimit nofile soft")
	flag.Parse()
}

func setupRlimits(hard, soft uint64) error {
	l := &syscall.Rlimit{
		Max: hard,
		Cur: soft,
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, l); err != nil {
		return errors.Wrap(err, "syscall.Setrlimit")
	}
	log.Infof("set ulimit nofile hard %d soft %d", l.Cur, l.Max)
	return nil
}

func main() {
	log.Infof("============= WOLF HOOK ==========")
	if err := setupRlimits(uint64(ulimitNofileHard), uint64(ulimitNofileSoft)); err != nil {
		log.Fatalf("setup ulimit nofile hard: %s", err)
	}

	go procutils.WaitZombieLoop(context.Background())

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
	r.Handle("/hook/start", handlers.NewStartController()).Methods("POST")
	r.Handle("/hook/stop", handlers.NewStopController()).Methods("POST")
	r.Handle("/hook/status", handlers.NewGetStatusController()).Methods("GET")
	r.Handle("/hook/exec", handlers.NewExecController()).Methods("POST")
	r.Handle("/hook/write-hwdb", handlers.NewWriteHwdbController()).Methods("POST")
	return r
}
