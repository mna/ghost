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

// Gzip compression HTTP handler.
type GzipHandler struct {
	H http.Handler
}

func (this *GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()

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

	// Do nothing on a HEAD request or if no accept-encoding is specified
	acc, ok := hdr["Accept-Encoding"]
	if r.Method == "HEAD" || !ok {
		this.H.ServeHTTP(w, r)
		return
	}

	// Check if gzip is an accepted response encoding
	ok = false
	for _, v := range acc {
		trimmed := strings.ToLower(strings.Trim(v, " "))
		if trimmed == "*" || trimmed == "gzip" {
			ok = true
			break
		}
	}
	if !ok {
		// No gzip support from the client, return uncompressed
		this.H.ServeHTTP(w, r)
		return
	}

	// Yes, prepare a gzip response container
	hdr.Set("Content-Encoding", "gzip")
	hdr.Del("Content-Length")
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Call the chained handler with a gzipped response writer
	this.H.ServeHTTP(
		&gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}, r)
}
