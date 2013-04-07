package ghost

import (
	"net/http"
)

// ChainableHandler is a valid Handler interface, and adds the possibility to
// chain other handlers.
type ChainableHandler interface {
	http.Handler
	Chain(http.Handler) ChainableHandler
}

type chainHandler struct {
	http.Handler
}

func (this *chainHandler) Chain(h http.Handler) ChainableHandler {
	return &chainHandler{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add the chained handler AFTER the call to this handler
			this.ServeHTTP(w, r)
			h.ServeHTTP(w, r)
		}),
	}
}

func NewChainableHandler(h http.Handler) ChainableHandler {
	return &chainHandler{
		h,
	}
}
