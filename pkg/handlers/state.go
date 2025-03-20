package handlers

import "sync"

type STATE string

const (
	STATE_RUNNING STATE = "RUNNING"
	STATE_STOPPED STATE = "STOPPED"
	STATE_ERROR   STATE = "ERROR"
)

var state STATE

var stateLock sync.Mutex

func SetState(s STATE) {
	stateLock.Lock()
	defer stateLock.Unlock()

	state = s
}

func GetState() STATE {
	return state
}
