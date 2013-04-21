package handlers

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func TestSessionExists(t *testing.T) {
	opts := NewSessionOptions(NewMemoryStore(1), "butchered at birth")
	h := SessionHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ssn, ok := GetSession(w)
			if assertTrue(ok, "expected session to be non-nil, got nil", t) {
				ssn.Data["foo"] = "bar"
				assertTrue(ssn.Data["foo"] == "bar", fmt.Sprintf("expected ssn[foo] to be 'bar', got %v", ssn.Data["foo"]), t)
			}
			w.Write([]byte("ok"))
		}), opts)
	s := httptest.NewServer(h)
	defer s.Close()

	var err error
	http.DefaultClient.Jar, err = cookiejar.New(new(cookiejar.Options))
	if err != nil {
		panic(err)
	}
	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 1, fmt.Sprintf("expected response to have 1 cookie, got %d", len(res.Cookies())), t)
}

func TestSessionPersists(t *testing.T) {
	cnt := 0
	opts := NewSessionOptions(NewMemoryStore(1), "butchered at birth")
	h := SessionHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ssn, ok := GetSession(w)
			if !ok {
				panic("session not found!")
			}
			if cnt == 0 {
				ssn.Data["foo"] = "bar"
				w.Write([]byte("ok"))
				cnt++
			} else {
				w.Write([]byte(ssn.Data["foo"].(string)))
			}
		}), opts)
	s := httptest.NewServer(h)
	defer s.Close()

	var err error
	http.DefaultClient.Jar, err = cookiejar.New(new(cookiejar.Options))
	if err != nil {
		panic(err)
	}
	// 1st call, set the session value
	res, err := http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)

	// 2nd call, get the session value
	t.Log("2nd call starting...")
	res, err = http.Get(s.URL)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("bar"), res, t)
	assertTrue(len(res.Cookies()) == 0, fmt.Sprintf("expected 2nd response to have 0 cookie, got %d", len(res.Cookies())), t)
}
