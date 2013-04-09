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
	assertStatus(http.StatusInternalServerError, res.StatusCode, t)
}

func TestNoPanic(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/nopanic", svrAddr)
	h := NewPanicHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

		}))
	startServer(h, "/nopanic")

	res, err := http.Get(path)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
}
