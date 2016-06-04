package main

import (
	"net/http"
	"net/url"
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

func getSiteURL(r *http.Request) string {
	if Config.siteURL != "" {
		return Config.siteURL
	} else {
		u := &url.URL{}
		u.Host = r.Host

		if Config.sitePath != "" {
			u.Path = Config.sitePath
		}

		if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
			u.Scheme = scheme
		} else if Config.certFile != "" || (r.TLS != nil && r.TLS.HandshakeComplete == true) {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}

		return u.String()
	}
}
