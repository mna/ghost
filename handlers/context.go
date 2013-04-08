package handlers

import (
	"net/http"
)

type contextResponseWriter struct {
	http.ResponseWriter
	m map[interface{}]interface{}
}

type ContextHandler struct {
	h       http.Handler
	initCap int
}

func NewContextHandler(wrappedHandler http.Handler, initialCap int) *ContextHandler {
	return &ContextHandler{
		wrappedHandler,
		initialCap,
	}
}

func (this *ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctxw := &contextResponseWriter{
		w,
		make(map[interface{}]interface{}, this.initCap),
	}
	this.h.ServeHTTP(ctxw, r)
}

func GetContext(w http.ResponseWriter) map[interface{}]interface{} {
	ctxw, ok := w.(*contextResponseWriter)
	if ok {
		return ctxw.m
	}
	return nil
}
