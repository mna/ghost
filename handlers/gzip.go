package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Thanks to Andrew Gerrand for inspiration:
// https://groups.google.com/d/msg/golang-nuts/eVnTcMwNVjM/4vYU8id9Q2UJ
//
// Also, node's Connect library implementation of the compress middleware:
// https://github.com/senchalabs/connect/blob/master/lib/middleware/compress.js
//
// And StackOverflow's explanation of Vary: Accept-Encoding header:
// http://stackoverflow.com/questions/7848796/what-does-varyaccept-encoding-mean

// Internal gzipped writer that satisfies both the (body) writer in gzipped format,
// and maintains the rest of the ResponseWriter interface for header manipulation.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Unambiguous Write() implementation (otherwise both ResponseWriter and Writer
// want to claim this method).
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Implement WrapWriter interface
func (w *gzipResponseWriter) WrappedWriter() http.ResponseWriter {
	return w.ResponseWriter
}

// Gzip compression HTTP handler.
func GZIPHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, ok := getGzipWriter(w); ok {
				// Self-awareness, gzip handler is already set up
				h.ServeHTTP(w, r)
				return
			}
			hdr := w.Header()
			setVaryHeader(hdr)

			// Do nothing on a HEAD request or if no accept-encoding is specified on the request
			acc, ok := r.Header["Accept-Encoding"]
			if r.Method == "HEAD" || !ok {
				h.ServeHTTP(w, r)
				return
			}
			if !acceptsGzip(acc) {
				// No gzip support from the client, return uncompressed
				h.ServeHTTP(w, r)
				return
			}

			// Prepare a gzip response container
			// TODO : Only if Content-Type is json/html/text?
			setGzipHeaders(hdr)
			gz := gzip.NewWriter(w)
			defer gz.Close()
			h.ServeHTTP(
				&gzipResponseWriter{
					Writer:         gz,
					ResponseWriter: w,
				}, r)
		})
}

// TODO : Generic header search function.
func setVaryHeader(hdr http.Header) {
	// Manage the Vary header field
	vary := hdr["Vary"]
	ok := false
	for _, v := range vary {
		if strings.ToLower(v) == "accept-encoding" {
			ok = true
		}
	}
	if !ok {
		hdr.Add("Vary", "Accept-Encoding")
	}
}

func acceptsGzip(acc []string) bool {
	for _, v := range acc {
		trimmed := strings.ToLower(strings.Trim(v, " "))
		if trimmed == "*" || strings.Contains(trimmed, "gzip") {
			return true
		}
	}
	return false
}

func setGzipHeaders(hdr http.Header) {
	// The content-type will be explicitly set somewhere down the path of handlers
	hdr.Set("Content-Encoding", "gzip")
	hdr.Del("Content-Length")
}

// Helper function to retrieve the gzip writer.
func getGzipWriter(w http.ResponseWriter) (*gzipResponseWriter, bool) {
	gz, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*gzipResponseWriter)
		return ok
	})
	if ok {
		return gz.(*gzipResponseWriter), true
	}
	return nil, false
}
