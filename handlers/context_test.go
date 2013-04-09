package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestContext(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/context", svrAddr)
	key := "key"
	val := 10
	body := "this is the output"

	// Inner handler to check the content of the context
	h2 := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := GetContext(w)
			v := ctx[key]
			assertTrue(v == val, fmt.Sprintf("expected value to be %v, got %v", val, v), t)

			// Actually write something
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
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
	startServer(h, "/context")

	// First call
	res, err := http.DefaultClient.Get(path)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
	// Second call, context should be cleaned at start
	res, err = http.DefaultClient.Get(path)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertBody([]byte(body), res, t)
}
