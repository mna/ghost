package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServeFile(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/servefile", svrAddr)
	h := StaticFileHandler("./testdata/styles.css")
	startServer(h, "/servefile")

	res, err := http.Get(path)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Type", "text/css; charset=utf-8", res, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(`* {
  background-color: white;
}`), res, t)
}

func TestGzippedFile(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/gzippedfile", svrAddr)
	h := GZIPHandler(StaticFileHandler("./testdata/styles.css"))
	startServer(h, "/gzippedfile")

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
	assertHeader("Content-Encoding", "gzip", res, t)
	assertHeader("Content-Type", "text/css; charset=utf-8", res, t)
	assertGzippedBody([]byte(`* {
  background-color: white;
}`), res, t)
}
