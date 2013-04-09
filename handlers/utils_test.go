package handlers

import (
	"net/http"
	"testing"
)

const svrAddr = ":8080"

var started = false

func startServer(h http.Handler, path string) {
	http.Handle(path, h)
	if !started {
		go http.ListenAndServe(svrAddr, nil)
		started = true
	}
}

func assertTrue(cond bool, msg string, t *testing.T) {
	if !cond {
		t.Error(msg)
	}
}
