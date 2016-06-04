package main

import (
	"net/http"
	"strings"
)

type addheaders struct {
	h       http.Handler
	headers []string
}

func (a addheaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, header := range a.headers {
		headerSplit := strings.SplitN(header, ": ", 2)
		w.Header().Add(headerSplit[0], headerSplit[1])
	}

	a.h.ServeHTTP(w, r)
}

func AddHeaders(headers []string) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return addheaders{h, headers}
	}
	return fn
}
