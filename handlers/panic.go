package handlers

import (
	"fmt"
	"net/http"
)

// Handles panics and responds with a 500 error message.
func PanicHandler(h http.Handler) http.Handler {
	// TODO : Add a custom message to override real error message if in production?
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
				}
			}()

			// Call the protected handler
			h.ServeHTTP(w, r)
		})
}
