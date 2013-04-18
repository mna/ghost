package handlers

import (
	"net/http"
	"strings"
)

const defaultCookieName = "ghost.sid"

type Session struct {
	ID   string
	Data map[interface{}]interface{}
}

type SessionOptions struct {
	Store          SessionStore
	CookieTemplate *http.Cookie
	TrustProxy     bool
}

type sessResponseWriter struct {
	http.ResponseWriter
	sess *Session
}

func (this *sessResponseWriter) WrappedWriter() http.ResponseWriter {
	return this.ResponseWriter
}

func SessionHandler(h http.Handler, opts *SessionOptions) http.Handler {
	ck := opts.CookieTemplate
	if ck == nil {
		ck = &http.Cookie{
			Name:   defaultCookieName,
			Path:   "/",
			MaxAge: 0,
		}
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, ok := GetSession(w); ok {
				// Self-awareness
				h.ServeHTTP(w, r)
				return
			}

			if strings.Index(r.URL.Path, ck.Path) != 0 {
				// Session cookie does not apply to this path
				h.ServeHTTP(w, r)
				return
			}
		})
}

// Helper function to retrieve the session-augmented writer.
func GetSession(w http.ResponseWriter) (*Session, bool) {
	ss, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*sessResponseWriter)
		return ok
	})
	if ok {
		return ss.(*sessResponseWriter).sess, true
	}
	return nil, false
}
