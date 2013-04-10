package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestLogging(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/logging", svrAddr)
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	log.SetFlags(0)

	h := NewLoggingHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

		}), nil, "%s!\n", "test")
	startServer(h, "/logging")

	res, err := http.Get(path)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertTrue(buf.String() == "ok!\n", fmt.Sprintf("expected log to be 'ok!', got %s", buf), t)
}
