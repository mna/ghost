package handlers

import (
	"net/http"
)

// Structure that holds the context map and exposes the ResponseWriter interface.
type contextResponseWriter struct {
	http.ResponseWriter
	m map[interface{}]interface{}
}

// ContextHandler gives a context storage that lives only for the duration of
// the request, with no locking involved.
func ContextHandler(h http.Handler, cap int) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, ok := w.(*contextResponseWriter); ok {
				// Self-awareness, context handler is already set up
				h.ServeHTTP(w, r)
				return
			}

			// Create the context-providing ResponseWriter replacement.
			ctxw := &contextResponseWriter{
				w,
				make(map[interface{}]interface{}, cap),
			}
			// Call the wrapped handler with the context-aware writer
			h.ServeHTTP(ctxw, r)
		})
}

// Helper function to retrieve the context map from the ResponseWriter interface.
func GetContext(w http.ResponseWriter) map[interface{}]interface{} {
	ctxw, ok := w.(*contextResponseWriter)
	if ok {
		return ctxw.m
	}
	return nil
}
