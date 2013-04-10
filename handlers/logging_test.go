package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestLogging(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/logging", svrAddr)
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	log.SetFlags(0)

	h := NewLoggingHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			t.Log("in handler")
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), nil, "%s - [%s] %s %s %s %d %s %s %s %.3f\n", "remote-addr", "date", "method", "url", "http-version", "status", "referrer", "user-agent", "bidon", "response-time")
	startServer(h, "/logging")

	res, err := http.Get(path)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertTrue(buf.String() == "ok!\n", fmt.Sprintf("expected log to be 'ok!', got %s", buf), t)
}
