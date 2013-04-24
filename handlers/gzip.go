package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
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
	encodingSet bool
	filterFn    func(http.ResponseWriter, *http.Request) bool
}

// Unambiguous Write() implementation (otherwise both ResponseWriter and Writer
// want to claim this method).
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// TODO : Filter...
	if !w.encodingSet {
		setGzipHeaders(w.Header())
		w.encodingSet = true
	}
	return w.Writer.Write(b)
}

// Intercept the WriteHeader call to correctly set the GZIP headers.
func (w *gzipResponseWriter) WriteHeader(code int) {
	// TODO : Filter...
	if !w.encodingSet {
		setGzipHeaders(w.Header())
		w.encodingSet = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Implement WrapWriter interface
func (w *gzipResponseWriter) WrappedWriter() http.ResponseWriter {
	return w.ResponseWriter
}

// Default filter to check if the response should be GZIPped.
// By default, all text (html, css, xml, ...), javascript and json
// content types are candidates for GZIP.
func defaultFilter(w http.ResponseWriter, r *http.Request) bool {
	hdr := w.Header()
	ok := HeaderMatch(hdr, "Content-Type", HmContains, "text")
	if !ok {
		ok = HeaderMatch(hdr, "Content-Type", HmContains, "javascript")
		if !ok {
			ok = HeaderMatch(hdr, "Content-Type", HmContains, "json")
		}
	}
	return ok
}

// GZIPHandlerFunc is the same as GZIPHandler, it is just a convenience
// signature that accepts a func(http.ResponseWriter, *http.Request) instead of
// a http.Handler interface. It saves the boilerplate http.HandlerFunc() cast.
func GZIPHandlerFunc(h http.HandlerFunc, filterFn func(http.ResponseWriter, *http.Request) bool) http.HandlerFunc {
	return GZIPHandler(h, filterFn)
}

// Gzip compression HTTP handler. If the client supports it, it compresses the response
// written by the wrapped handler. The filter function is called when the response is about
// to be written to determine if compression should be applied. If this argument is nil,
// the default filter will GZIP only content types containing /json|text|javascript/.
func GZIPHandler(h http.Handler, filterFn func(http.ResponseWriter, *http.Request) bool) http.HandlerFunc {
	if filterFn == nil {
		filterFn = defaultFilter
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getGzipWriter(w); ok {
			// Self-awareness, gzip handler is already set up
			h.ServeHTTP(w, r)
			return
		}
		hdr := w.Header()
		setVaryHeader(hdr)

		// Do nothing on a HEAD request
		if r.Method == "HEAD" {
			h.ServeHTTP(w, r)
			return
		}
		if !acceptsGzip(r.Header) {
			// No gzip support from the client, return uncompressed
			h.ServeHTTP(w, r)
			return
		}

		// Prepare a gzip response container
		gz := gzip.NewWriter(w)
		h.ServeHTTP(
			&gzipResponseWriter{
				Writer:         gz,
				ResponseWriter: w,
			}, r)
		// Iff the handler completed successfully (no panic), close the gzip writer,
		// which seems to generate a Write to the underlying writer.
		gz.Close()
	}
}

// Add the vary by "accept-encoding" header if it is not already set.
func setVaryHeader(hdr http.Header) {
	if !HeaderMatch(hdr, "Vary", HmContains, "accept-encoding") {
		hdr.Add("Vary", "Accept-Encoding")
	}
}

// Checks if the client accepts GZIP-encoded responses.
func acceptsGzip(hdr http.Header) bool {
	ok := HeaderMatch(hdr, "Accept-Encoding", HmContains, "gzip")
	if !ok {
		ok = HeaderMatch(hdr, "Accept-Encoding", HmEquals, "*")
	}
	return ok
}

func setGzipHeaders(hdr http.Header) {
	// The content-type will be explicitly set somewhere down the path of handlers
	hdr.Set("Content-Encoding", "gzip")
	// BUG : No clever way to set the content length...
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
