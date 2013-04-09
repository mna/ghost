package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestContext(t *testing.T) {
	key := "key"
	val := 10

	// Inner handler to check the content of the context
	h2 := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := GetContext(w)
			v := ctx[key]
			assertTrue(v == val, fmt.Sprintf("expected value to be %v, got %v", val, v), t)
		})

	// Create the context handler with a wrapped handler
	h := NewContextHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := GetContext(w)
			assertTrue(ctx != nil, "expected context to be non-nil", t)
			assertTrue(len(ctx) == 0, fmt.Sprintf("expected context to be empty, got %d", len(ctx)), t)
			ctx[key] = val
			h2.ServeHTTP(w, r)
		}), 2)

	// Start and stop the server
	l := startServer(h, t)
	defer l.Close()

	// First call
	_, err := http.DefaultClient.Get(clientAddr)
	if err != nil {
		panic(err)
	}
	// Second call, context should be cleaned at start
	_, err = http.DefaultClient.Get(clientAddr)
	if err != nil {
		panic(err)
	}
}
