package handlers

import (
	"log"
	"net/http"
	"time"
)

// TODO : Predefined formats

type LoggingHandler struct {
	H            http.Handler
	Logger       *log.Logger
	Format       string
	Tokens       []string
	CustomTokens map[string]func(http.ResponseWriter, *http.Request) string
	Immediate    bool
}

func NewLoggingHandler(wrappedHandler http.Handler, l *log.Logger, fmt string, tok ...string) *LoggingHandler {
	return &LoggingHandler{
		wrappedHandler,
		l,
		fmt,
		tok,
		make(map[string]func(http.ResponseWriter, *http.Request) string),
		false,
	}
}

func (this *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Save the response start time
	st := time.Now()
	// Log immediately if requested, otherwise on exit
	if this.Immediate {
		this.log(w, r, st)
	} else {
		defer this.log(w, r, st)
	}
	// Call the wrapped handler
	this.H.ServeHTTP(w, r)
}

func (this *LoggingHandler) getPredefinedTokenValue(t string, w http.ResponseWriter,
	r *http.Request, st time.Time) (string, bool) {
	return "ok", true
}

func (this *LoggingHandler) log(w http.ResponseWriter, r *http.Request, st time.Time) {
	var fn func(string, ...interface{})
	var ok bool

	if this.Logger == nil {
		fn = log.Printf
	} else {
		fn = this.Logger.Printf
	}
	args := make([]interface{}, len(this.Tokens))
	for i, t := range this.Tokens {
		if args[i], ok = this.getPredefinedTokenValue(t, w, r, st); !ok {
			if args[i], ok = this.CustomTokens[t]; !ok {
				args[i] = "?"
			}
		}
	}
	fn(this.Format, args...)
}
