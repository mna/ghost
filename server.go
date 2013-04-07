package ghost

import (
	"net/http"
)

type Server struct {
	*http.Server
}

func (this *Server) Run() {
	this.ListenAndServe()
}
