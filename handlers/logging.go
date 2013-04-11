package handlers

// Inspired by node's Connect library implementation of the logging middleware
// https://github.com/senchalabs/connect

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

const (
	// Predefined logging formats that can be passed as format string.
	Ldefault = "_default_"
	Lshort   = "_short_"
	Ltiny    = "_tiny_"
)

// Token for request and response headers
var (
	rxHeaders = regexp.MustCompile(`^(req|res)\[([^\]]+)\]$`)

	predefFormats = map[string]struct {
		fmt  string
		toks []string
	}{
		Ldefault: {
			`%s - - [%s] "%s %s HTTP/%s" %d %s "%s" "%s"`,
			[]string{"remote-addr", "date", "method", "url", "http-version", "status", "res[Content-Length]", "referrer", "user-agent"},
		},
		Lshort: {
			`%s - %s %s HTTP/%s %d %s - %.3f s`,
			[]string{"remote-addr", "method", "url", "http-version", "status", "res[Content-Length]", "response-time"},
		},
		Ltiny: {
			`%s %s %d %s - %.3f s`,
			[]string{"method", "url", "status", "res[Content-Length]", "response-time"},
		},
	}
)

// Augmented ResponseWriter implementation that captures the status code for the logger.
type statusResponseWriter struct {
	// TODO : This solution only works if WriteHeader() is called explicitly.
	// If Write() is called and then this Write calls WriteHeader, the WriteHeader
	// of the internal *response struct is called, *not* the overridden method of
	// this statusResponseWriter struct.
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
	Immediate    bool   // Defaults to false
	DateFormat   string // Defaults to time.RFC3339
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
		time.RFC3339,
	}
}

// Implement the http.Handler interface.
func (this *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(*statusResponseWriter); ok {
		// Self-awareness, logging handler already set up
		this.H.ServeHTTP(w, r)
		return
	}

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
		return fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor), true
	case "response-time":
		return time.Now().Sub(st).Seconds(), true
	case "remote-addr":
		return r.RemoteAddr, true
	case "date":
		return time.Now().Format(this.DateFormat), true
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

	// Handle special cases for header
	mtch := rxHeaders.FindStringSubmatch(t)
	if len(mtch) > 2 {
		if mtch[1] == "req" {
			return r.Header.Get(mtch[2]), true
		} else {
			// TODO : This only works for headers explicitly set via the Header() map of
			// the writer, not those added by the http package under the covers.
			return w.Header().Get(mtch[2]), true
		}
	}
	return nil, false
}

// Do the actual logging.
func (this *LoggingHandler) log(w http.ResponseWriter, r *http.Request, st time.Time) {
	var fn func(string, ...interface{})
	var ok bool
	var format string
	var toks []string

	// If no specific logger, use the default one from the log package
	if this.Logger == nil {
		fn = log.Printf
	} else {
		fn = this.Logger.Printf
	}

	// If this is a predefined format, use it instead
	if v, ok := predefFormats[this.Format]; ok {
		format = v.fmt
		toks = v.toks
	} else {
		format = this.Format
		toks = this.Tokens
	}
	args := make([]interface{}, len(toks))
	for i, t := range toks {
		if args[i], ok = this.getPredefinedTokenValue(t, w, r, st); !ok {
			if f, ok := this.CustomTokens[t]; ok && f != nil {
				args[i] = f(w, r)
			} else {
				args[i] = "?"
			}
		}
	}
	fn(format, args...)
}
