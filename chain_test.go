package ghost

import (
	"bytes"
	"net/http"
	"testing"
)

func TestChaining(t *testing.T) {
	var buf bytes.Buffer

	a := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('a')
	}
	b := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('b')
	}
	c := func(w http.ResponseWriter, r *http.Request) {
		buf.WriteRune('c')
	}
	f := NewChainableHandler(http.HandlerFunc(a)).Chain(http.HandlerFunc(b)).Chain(http.HandlerFunc(c))
	f.ServeHTTP(nil, nil)

	if buf.String() != "abc" {
		t.Errorf("expected 'abc', got %s", buf.String())
	}
}
