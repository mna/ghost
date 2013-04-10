package handlers

import (
	"log"
	"net/http"
	"time"
)

// Inspired by node's connect library logging middleware
// https://github.com/senchalabs/connect

// TODO : Predefined formats

// Augmented ResponseWriter implementation that captures the status code for the logger.
type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

// Intercept the WriteHeader call to save the status code.
func (this *statusResponseWriter) WriteHeader(code int) {
	this.code = code
	this.ResponseWriter.WriteHeader(code)
}

// Generic, customizable logging handler.
type LoggingHandler struct {
	H            http.Handler
	Logger       *log.Logger
	Format       string
	Tokens       []string
	CustomTokens map[string]func(http.ResponseWriter, *http.Request) string
	Immediate    bool
}

// Create a new logging handler for the specified wrapped handler and other logging options.
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

// Implement the http.Handler interface.
func (this *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Save the response start time
	st := time.Now()
	// Call the wrapped handler, with the augmented ResponseWriter to handle the status code
	stw := &statusResponseWriter{w, 0}

	// Log immediately if requested, otherwise on exit
	if this.Immediate {
		this.log(w, r, st)
	} else {
		defer this.log(stw, r, st)
	}
	this.H.ServeHTTP(stw, r)
}

// Check if the specified token is a predefined one, and if so return its current value.
func (this *LoggingHandler) getPredefinedTokenValue(t string, w http.ResponseWriter,
	r *http.Request, st time.Time) (interface{}, bool) {
	switch t {
	case "http-version":
		return r.Proto, true
	case "response-time":
		return time.Now().Sub(st).Seconds(), true
	case "remote-addr":
		return r.RemoteAddr, true
	case "date":
		return time.Now(), true
	case "method":
		return r.Method, true
	case "url":
		return r.URL.String(), true
	case "referrer", "referer":
		return r.Referer(), true
	case "user-agent":
		return r.UserAgent(), true
	case "status":
		if stw, ok := w.(*statusResponseWriter); ok && stw.code > 0 {
			return stw.code, true
		}
	}
	return nil, false
}

// Do the actual logging.
func (this *LoggingHandler) log(w http.ResponseWriter, r *http.Request, st time.Time) {
	var fn func(string, ...interface{})
	var ok bool

	// If no specific logger, use the default one from the log package
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
