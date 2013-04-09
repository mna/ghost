package handlers

import (
	"bytes"
	"io/ioutil"
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

func assertStatus(ex, ac int, t *testing.T) {
	if ex != ac {
		t.Errorf("expected status code to be %d, got %d", ex, ac)
	}
}

func assertBody(ex []byte, res *http.Response, t *testing.T) {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if !bytes.Equal(ex, buf) {
		t.Errorf("expected body to be '%v' (%d), got '%v' (%d)", ex, len(ex), buf, len(buf))
	}
}
