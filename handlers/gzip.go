package handlers

import (
	"net/http"
)

// Thanks to Andrew Gerrand for inspiration
// https://groups.google.com/d/msg/golang-nuts/eVnTcMwNVjM/4vYU8id9Q2UJ
type GzipHandler struct {
	next http.Handler
}

func (this *GzipHandler) Chain(http.Handler) ChainableHandler {

}
