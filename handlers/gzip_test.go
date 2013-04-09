package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGzipped(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/gzipped", svrAddr)
	body := "This is the body"
	headers := []string{"gzip", "*", "gzip, deflate, sdch"}

	h := NewGzipHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}))
	startServer(h, "/gzipped")

	for _, hdr := range headers {
		t.Logf("running with Accept-Encoding header %s", hdr)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Accept-Encoding", hdr)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		assertStatus(http.StatusOK, res.StatusCode, t)
		assertHeader("Content-Encoding", "gzip", res, t)
		assertGzippedBody([]byte(body), res, t)
	}
}

func TestNoGzip(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/nogzip", svrAddr)
	body := "This is the body"

	h := NewGzipHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}))
	startServer(h, "/nogzip")

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}
