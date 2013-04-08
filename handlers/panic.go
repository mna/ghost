package handlers

import (
	"fmt"
	"net/http"
)

// Handles panics and responds with a 500 error message.
type PanicHandler struct{}

// Implementation of the http.Handler interface.
func (this *PanicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
		}
	}()
}
