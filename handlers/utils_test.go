package handlers

import (
	"fmt"
	"net"
	"net/http"
	"testing"
)

const svrAddr = ":8080"

var clientAddr = fmt.Sprintf("http://localhost%s/", svrAddr)

func startServer(h http.Handler, t *testing.T) net.Listener {
	l, err := net.Listen("tcp", svrAddr)
	if err != nil {
		panic(err)
	}
	go http.Serve(l, h)

	return l
}

func assertTrue(cond bool, msg string, t *testing.T) {
	if !cond {
		t.Error(msg)
	}
}
