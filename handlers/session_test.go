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
}
