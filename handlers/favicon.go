package handlers

import (
	"github.com/PuerkitoBio/ghost"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// FaviconHandlerFunc is the same as FaviconHandler, it is just a convenience
// signature that accepts a func(http.ResponseWriter, *http.Request) instead of
// a http.Handler interface. It saves the boilerplate http.HandlerFunc() cast.
func FaviconHandlerFunc(h http.HandlerFunc, path string, maxAge time.Duration) http.HandlerFunc {
	return FaviconHandler(h, path, maxAge)
}

// Efficient favicon handler, mostly a port of node's Connect library implementation
// of the favicon middleware.
// https://github.com/senchalabs/connect
func FaviconHandler(h http.Handler, path string, maxAge time.Duration) http.HandlerFunc {
	var buf []byte

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		if r.URL.Path == "/favicon.ico" {
			if buf != nil {
				// Serve from cache
				writeHeaders(w.Header(), buf, maxAge)
				writeBody(w, r, buf)
			} else {
				// Read from file and cache
				buf, err = ioutil.ReadFile(path)
				if err != nil {
					ghost.LogFn("ghost.favicon : error reading file : %s", err)
					http.NotFound(w, r)
					return
				}
				writeHeaders(w.Header(), buf, maxAge)
				writeBody(w, r, buf)
			}
		} else {
			h.ServeHTTP(w, r)
		}
	}
}

// Write the content of the favicon, or respond with a 404 not found
// in case of error (hardly a critical error).
func writeBody(w http.ResponseWriter, r *http.Request, buf []byte) {
	_, err := w.Write(buf)
	if err != nil {
		ghost.LogFn("ghost.favicon : error writing response : %s", err)
		http.NotFound(w, r)
	}
}

func writeHeaders(hdr http.Header, buf []byte, maxAge time.Duration) {
	hdr.Set("Content-Type", "image/x-icon")
	hdr.Set("Content-Length", strconv.Itoa(len(buf)))
	hdr.Set("Etag", "") // TODO : MD5 hash, cached
	hdr.Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds())))
}
