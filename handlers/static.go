package handlers

import (
	"net/http"
)

// StaticFileHandler, unlike net/http.FileServer, serves the contents of a specific
// file when it is called.
func StaticFileHandler(path string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path)
		})
}
