package main

import (
	"net/http"
)

const (
	cspHeader          = "Content-Security-Policy"
	rpHeader           = "Referrer-Policy"
	frameOptionsHeader = "X-Frame-Options"
)

type csp struct {
	h    http.Handler
	opts CSPOptions
}

type CSPOptions struct {
	policy         string
	referrerPolicy string
	frame          string
}

func (c csp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only add a CSP if one is not already set
	if existing := w.Header().Get(cspHeader); existing == "" {
		w.Header().Add(cspHeader, c.opts.policy)
	}

	// only add a Referrer Policy if one is not already set
	if existing := w.Header().Get(rpHeader); existing == "" {
		w.Header().Add(rpHeader, c.opts.referrerPolicy)
	}

	w.Header().Set(frameOptionsHeader, c.opts.frame)

	c.h.ServeHTTP(w, r)
}

func ContentSecurityPolicy(o CSPOptions) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return csp{h, o}
	}
	return fn
}
