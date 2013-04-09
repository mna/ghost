package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"testing"
)

func TestGzipped(t *testing.T) {
	// TODO : Validate result. Maybe response's body is already unzipped
	path := fmt.Sprintf("http://localhost%s/gzipped", svrAddr)
	body := "This is the body"
	gbody := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(gbody)
	if _, err := gw.Write([]byte(body)); err != nil {
		panic(err)
	}
	h := NewGzipHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}))
	startServer(h, "/gzipped")

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "*")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody(gbody.Bytes(), res, t)
}
