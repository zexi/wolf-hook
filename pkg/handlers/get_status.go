package handlers

import "net/http"

type getStatusController struct{}

func NewGetStatusController() http.Handler {
	return new(getStatusController)
}

func (g getStatusController) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(GetState()))
	return
}
