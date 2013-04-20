package handlers

import (
	"encoding/json"
	"errors"
	"hash/crc32"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/ghost"
	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"
)

const defaultCookieName = "ghost.sid"

var (
	ErrSessionSecretMissing = errors.New("session secret is missing")
)

type Session struct {
	ID   string
	Data map[interface{}]interface{}
	// TODO : If MaxAge or Expires field, flag as json-ignore or internal fields?
	maxAge         int
	originalMaxAge int
}

func newSession() *Session {
	uid, err := uuid.NewV4()
	if err != nil {
		ghost.LogFn("ghost.session : error generating session ID : %s", err)
		return nil
	}
	return &Session{
		uid.String(),
		make(map[interface{}]interface{}),
		0,
		0,
	}
}

func (this *Session) resetMaxAge() {
	this.maxAge = this.originalMaxAge
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
		ghost.LogFn("ghost.session : secure cookie on a non-secure connection, cookie not sent")
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
			var sess *Session
			exCk, err := r.Cookie(opts.CookieTemplate.Name)
			if err != nil {
				sess = newSession()
				ghost.LogFn("ghost.session : error getting session cookie : %s", err)
			} else {
				ckSessId, err := parseSignedCookie(exCk, opts.Secret)
				if err != nil {
					sess = newSession()
					ghost.LogFn("ghost.session : error parsing signed cookie : %s", err)
				} else if ckSessId == "" {
					sess = newSession()
					ghost.LogFn("ghost.session : no existing session ID")
				} else {
					// Get the session
					sess, err := opts.Store.Get(ckSessId)
					if err != nil {
						sess = newSession()
						ghost.LogFn("ghost.session : error getting session from store : %s", err)
					} else if sess == nil {
						sess = newSession()
						ghost.LogFn("ghost.session : nil session")
					}
				}
			}
			_ = hash(sess)
			srw := &sessResponseWriter{w, nil, false, &opts, r, ""}
			defer func() {
				srw.sess.resetMaxAge()
				err := srw.opts.Store.Set(srw.sess.ID, srw.sess)
				if err != nil {
					ghost.LogFn("ghost.session : error saving session to store : %s", err)
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

func hash(s *Session) uint32 {
	data, err := json.Marshal(s)
	if err != nil {
		return 0 // TODO : Return what? 0 means treat as different session value?
	}
	return crc32.ChecksumIEEE(data)
}
