package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnauth(t *testing.T) {
	h := BasicAuthHandler(StaticFileHandler("./testdata/script.js"), func(u, pwd string) (interface{}, bool) {
		if u == "me" && pwd == "you" {
			return u, true
		}
		return nil, false
	}, "foo")
	s := httptest.NewServer(h)
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusUnauthorized, res.StatusCode, t)
	assertHeader("Www-Authenticate", `Basic realm="foo"`, res, t)
}

func TestGzippedAuth(t *testing.T) {
	h := GZIPHandler(BasicAuthHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			usr, ok := GetUser(w)
			if assertTrue(ok, "expected authenticated user, got false", t) {
				assertTrue(usr.(string) == "me", fmt.Sprintf("expected user to be 'me', got '%s'", usr), t)
			}
			w.Write([]byte(usr.(string)))
		}), func(u, pwd string) (interface{}, bool) {
		if u == "me" && pwd == "you" {
			return u, true
		}
		return nil, false
	}, ""))

	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", "http://me:you@"+s.URL[7:], nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertGzippedBody([]byte("me"), res, t)
}
