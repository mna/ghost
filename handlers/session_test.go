package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	memStore = NewMemoryStore(1)
	secret   = "butchered at birth"
)

func setupTest(f func(w http.ResponseWriter, r *http.Request), ckPath string, secure bool, maxAge int) *httptest.Server {
	opts := NewSessionOptions(memStore, secret)
	if ckPath != "" {
		opts.CookieTemplate.Path = ckPath
	}
	opts.CookieTemplate.Secure = secure
	opts.CookieTemplate.MaxAge = maxAge
	h := SessionHandler(http.HandlerFunc(f), opts)
	return httptest.NewServer(h)
}

func doRequest(u string, newJar bool) *http.Response {
	var err error
	if newJar {
		http.DefaultClient.Jar, err = cookiejar.New(new(cookiejar.Options))
		if err != nil {
			panic(err)
		}
	}
	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	return res
}

func TestSessionExists(t *testing.T) {
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		ssn, ok := GetSession(w)
		if assertTrue(ok, "expected session to be non-nil, got nil", t) {
			ssn.Data["foo"] = "bar"
			assertTrue(ssn.Data["foo"] == "bar", fmt.Sprintf("expected ssn[foo] to be 'bar', got %v", ssn.Data["foo"]), t)
		}
		w.Write([]byte("ok"))
	}, "", false, 0)
	defer s.Close()

	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 1, fmt.Sprintf("expected response to have 1 cookie, got %d", len(res.Cookies())), t)
}

func TestSessionPersists(t *testing.T) {
	cnt := 0
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
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
	}, "", false, 0)
	defer s.Close()

	// 1st call, set the session value
	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)

	// 2nd call, get the session value
	res = doRequest(s.URL, false)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("bar"), res, t)
	assertTrue(len(res.Cookies()) == 0, fmt.Sprintf("expected 2nd response to have 0 cookie, got %d", len(res.Cookies())), t)
}

func TestSessionExpires(t *testing.T) {
	cnt := 0
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		ssn, ok := GetSession(w)
		if !ok {
			panic("session not found!")
		}
		if cnt == 0 {
			w.Write([]byte(ssn.ID()))
			cnt++
		} else {
			w.Write([]byte(ssn.ID()))
		}
	}, "", false, 1) // Expire in 1 second
	defer s.Close()

	// 1st call, set the session value
	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	id1, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	time.Sleep(1001 * time.Millisecond)

	// 2nd call, get the session value
	res = doRequest(s.URL, false)
	assertStatus(http.StatusOK, res.StatusCode, t)
	id2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	sid1, sid2 := string(id1), string(id2)
	assertTrue(len(res.Cookies()) == 1, fmt.Sprintf("expected 2nd response to have 1 cookie, got %d", len(res.Cookies())), t)
	assertTrue(sid1 != sid2, "expected session IDs to be different, got same", t)
}

func TestSessionBeforeExpires(t *testing.T) {
	cnt := 0
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		ssn, ok := GetSession(w)
		if !ok {
			panic("session not found!")
		}
		if cnt == 0 {
			w.Write([]byte(ssn.ID()))
			cnt++
		} else {
			w.Write([]byte(ssn.ID()))
		}
	}, "", false, 1) // Expire in 1 second
	defer s.Close()

	// 1st call, set the session value
	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	id1, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// 2nd call, get the session value
	res = doRequest(s.URL, false)
	assertStatus(http.StatusOK, res.StatusCode, t)
	id2, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	sid1, sid2 := string(id1), string(id2)
	assertTrue(len(res.Cookies()) == 0, fmt.Sprintf("expected 2nd response to have no cookie, got %d", len(res.Cookies())), t)
	assertTrue(sid1 == sid2, "expected session IDs to be the same, got different", t)
}

func TestPanicIfNoSecret(t *testing.T) {
	defer assertPanic(t)
	SessionHandler(http.NotFoundHandler(), NewSessionOptions(nil, ""))
}

func TestInvalidPath(t *testing.T) {
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSession(w)
		assertTrue(!ok, "expected session to be nil, got non-nil", t)
		w.Write([]byte("ok"))
	}, "/foo", false, 0)
	defer s.Close()

	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 0, fmt.Sprintf("expected response to have no cookie, got %d", len(res.Cookies())), t)
}

func TestValidSubPath(t *testing.T) {
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSession(w)
		assertTrue(ok, "expected session to be non-nil, got nil", t)
		w.Write([]byte("ok"))
	}, "/foo", false, 0)
	defer s.Close()

	res := doRequest(s.URL+"/foo/bar", true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 1, fmt.Sprintf("expected response to have 1 cookie, got %d", len(res.Cookies())), t)
}

func TestSecureOverHttp(t *testing.T) {
	s := setupTest(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetSession(w)
		assertTrue(ok, "expected session to be non-nil, got nil", t)
		w.Write([]byte("ok"))
	}, "", true, 0)
	defer s.Close()

	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 0, fmt.Sprintf("expected response to have no cookie, got %d", len(res.Cookies())), t)
}

// TODO : commented, certificate problem
func xTestSecureOverHttps(t *testing.T) {
	opts := NewSessionOptions(memStore, secret)
	opts.CookieTemplate.Secure = true
	h := SessionHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetSession(w)
			assertTrue(ok, "expected session to be non-nil, got nil", t)
			w.Write([]byte("ok"))
		}), opts)
	s := httptest.NewTLSServer(h)
	defer s.Close()

	res := doRequest(s.URL, true)
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte("ok"), res, t)
	assertTrue(len(res.Cookies()) == 1, fmt.Sprintf("expected response to have 1 cookie, got %d", len(res.Cookies())), t)
}
