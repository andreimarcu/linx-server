package main

import (
	"net/http"
)

const (
	cspHeader                = "Content-Security-Policy"
	frameOptionsHeader       = "X-Frame-Options"
	contentTypeOptionsHeader = "X-Content-Type-Options"
)

type csp struct {
	h    http.Handler
	opts CSPOptions
}

type CSPOptions struct {
	policy string
	frame  string
}

func (c csp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only add a CSP if one is not already set
	if existing := w.Header().Get(cspHeader); existing == "" {
		w.Header().Add(cspHeader, c.opts.policy)
	}

	w.Header().Set(frameOptionsHeader, c.opts.frame)
	w.Header().Set(contentTypeOptionsHeader, "nosniff")

	c.h.ServeHTTP(w, r)
}

func ContentSecurityPolicy(o CSPOptions) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return csp{h, o}
	}
	return fn
}

// vim:set ts=8 sw=8 noet:
