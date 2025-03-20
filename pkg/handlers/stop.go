package handlers

import (
	"net/http"
	"os"
	"time"

	"yunion.io/x/log"
)

type stopController struct{}

func NewStopController() http.Handler {
	return new(stopController)
}

func (s *stopController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Stop request: %s", r.URL.Path)
	SetState(STATE_STOPPED)
	w.WriteHeader(http.StatusAccepted)
	go func() {
		log.Infof("System will be stopped after 2 seconds")
		time.Sleep(2 * time.Second)
		log.Infof("System stopped")
		os.Exit(0)
	}()
}
