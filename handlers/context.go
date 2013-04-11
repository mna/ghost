package handlers

import (
	"net/http"
)

// Structure that holds the context map and exposes the ResponseWriter interface.
type contextResponseWriter struct {
	http.ResponseWriter
	m map[interface{}]interface{}
}

// The ContextHandler gives a context storage that lives only for the duration of
// the request, with no locking involved.
type ContextHandler struct {
	H          http.Handler
	InitialCap int
}

// Create a new context handler for the specified wrapped handler, and an initial 
// capacity for the context map.
func NewContextHandler(wrappedHandler http.Handler) *ContextHandler {
	return &ContextHandler{
		H: wrappedHandler,
	}
}

// Implementation of the http.Handler interface.
func (this *ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(*contextResponseWriter); ok {
		// Self-awareness, context handler is already set up
		this.H.ServeHTTP(w, r)
		return
	}

	// Create the context-providing ResponseWriter replacement.
	ctxw := &contextResponseWriter{
		w,
		make(map[interface{}]interface{}, this.InitialCap),
	}
	// Call the wrapped handler with the context-aware writer
	this.H.ServeHTTP(ctxw, r)
}

// Helper function to retrieve the context map from the ResponseWriter interface.
func GetContext(w http.ResponseWriter) map[interface{}]interface{} {
	ctxw, ok := w.(*contextResponseWriter)
	if ok {
		return ctxw.m
	}
	return nil
}
