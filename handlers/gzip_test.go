package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipped(t *testing.T) {
	body := "This is the body"
	headers := []string{"gzip", "*", "gzip, deflate, sdch"}

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}))
	s := httptest.NewServer(h)
	defer s.Close()

	for _, hdr := range headers {
		t.Logf("running with Accept-Encoding header %s", hdr)
		req, err := http.NewRequest("GET", s.URL, nil)
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
	body := "This is the body"

	h := GZIPHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}))
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
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
