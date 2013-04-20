package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
)

const defaultCookieName = "ghost.sid"

var (
	ErrSessionSecretMissing = errors.New("session secret is missing")
)

type Session struct {
	ID   string
	Data map[interface{}]interface{}
}

type SessionOptions struct {
	Store          SessionStore
	CookieTemplate http.Cookie
	TrustProxy     bool
	Secret         string
}

type sessResponseWriter struct {
	http.ResponseWriter
	sess         *Session
	sessSent     bool
	opts         *SessionOptions
	req          *http.Request
	cookieSessID string
}

func (this *sessResponseWriter) WrappedWriter() http.ResponseWriter {
	return this.ResponseWriter
}

func (this *sessResponseWriter) Write(data []byte) (int, error) {
	if !this.sessSent {
		this.sendSessionCookie()
		this.sessSent = true
	}
	return this.ResponseWriter.Write(data)
}

func (this *sessResponseWriter) WriteHeader(code int) {
	if !this.sessSent {
		this.sendSessionCookie()
		this.sessSent = true
	}
	this.ResponseWriter.WriteHeader(code)

}

func (this *sessResponseWriter) sendSessionCookie() {
	if this.sess == nil {
		return
	}
	proto := strings.Trim(strings.ToLower(this.req.Header.Get("X-Forwarded-Proto")), " ")
	tls := this.req.TLS != nil || (strings.HasPrefix(proto, "https") && this.opts.TrustProxy)
	if this.opts.CookieTemplate.Secure && !tls {
		// TODO : Log
		// Requested secure cookie, but not a secure connection, do not send
		return
	}
	isNew := this.cookieSessID != this.sess.ID
	if !isNew {
		if this.opts.CookieTemplate.RawExpires == "" {

		}
	}
}

func SessionHandler(h http.Handler, opts SessionOptions) http.Handler {
	// Make sure the required cookie fields are set
	if opts.CookieTemplate.Name == "" {
		opts.CookieTemplate.Name = defaultCookieName
	}
	if opts.CookieTemplate.Path == "" {
		opts.CookieTemplate.Path = "/"
	}
	// Secret is required
	if opts.Secret == "" {
		panic(ErrSessionSecretMissing)
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, ok := getSessionWriter(w); ok {
				// Self-awareness
				h.ServeHTTP(w, r)
				return
			}

			if strings.Index(r.URL.Path, opts.CookieTemplate.Path) != 0 {
				// Session cookie does not apply to this path
				h.ServeHTTP(w, r)
				return
			}
			// Get the session cookie
			exCk, err := r.Cookie(opts.CookieTemplate.Name)
			if err != nil {
				// TODO : Generate a new Session
			}
			ckSessId, err := parseSignedCookie(exCk, opts.Secret)
			if err != nil {
				// TODO : Generate a new session if none yet
			}
			if ckSessId == "" {
				// TODO : Generate a new session if none yet
			}
			// Get the session
			sess, err := opts.Store.Get(ckSessId)
			if err != nil {
				// TODO : Generate a new session if none yet
			} else if sess == nil {
				// TODO : Generate a new session if none yet
			}
			srw := &sessResponseWriter{w, nil, false, &opts, r, ckSessId}
			defer func() {
				srw.sess.resetMaxAge()
				err := srw.opts.Store.Set(srw.sess.ID, srw.sess)
				if err != nil {
					// TODO : Log error
				}
			}()
			h.ServeHTTP(srw, r)
		})
}

// Helper function to retrieve the session for the current request.
func GetSession(w http.ResponseWriter) (*Session, bool) {
	ss, ok := getSessionWriter(w)
	if ok {
		return ss.sess, true
	}
	return nil, false
}

// Helper function to retrieve the session store
func GetSessionStore(w http.ResponseWriter) (SessionStore, bool) {
	ss, ok := getSessionWriter(w)
	if ok {
		return ss.opts.Store, true
	}
	return nil, false
}

// Internal helper function to retrieve the session writer object.
func getSessionWriter(w http.ResponseWriter) (*sessResponseWriter, bool) {
	ss, ok := GetResponseWriter(w, func(tst http.ResponseWriter) bool {
		_, ok := tst.(*sessResponseWriter)
		return ok
	})
	if ok {
		return ss.(*sessResponseWriter), true
	}
	return nil, false
}

func parseSignedCookie(ck *http.Cookie, secret string) (string, error) {
	var val string

	sck := securecookie.New([]byte(secret), nil)
	err := sck.Decode(ck.Name, ck.Value, val)
	if err != nil {
		return "", err
	}
	return val, nil
}
