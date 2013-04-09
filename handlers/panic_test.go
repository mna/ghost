package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPanic(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/panic", svrAddr)
	h := NewPanicHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("test")
		}))
	startServer(h, "/panic")

	res, err := http.Get(path)
	if err != nil {
		panic(err)
	}
	assertTrue(res.StatusCode == http.StatusInternalServerError, fmt.Sprintf("expected status code to be 500, got %d", res.StatusCode), t)
}
