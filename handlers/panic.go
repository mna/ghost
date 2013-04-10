package handlers

import (
	"fmt"
	"net/http"
)

// Handles panics and responds with a 500 error message.
type PanicHandler struct {
	H http.Handler
	// TODO : Add a custom message to override real error message if in production?
}

// Create a new panic handler around a handler, making it a "protected" handler
// (panics will result in a 500 response).
func NewPanicHandler(protectedHandler http.Handler) *PanicHandler {
	return &PanicHandler{protectedHandler}
}

// Implementation of the http.Handler interface.
func (this *PanicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		}
	}()

	// Call the protected handler
	this.H.ServeHTTP(w, r)
}
