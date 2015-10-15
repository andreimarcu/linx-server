package main

import (
	"net/http"
	"net/url"
)

// Do a strict referrer check, matching against both the Origin header (if
// present) and the Referrer header. If a list of headers is specified, then
// Referrer checking will be skipped if any of those headers are present.
func strictReferrerCheck(r *http.Request, prefix string, whitelistHeaders []string) bool {
	p, _ := url.Parse(prefix)

	// if there's an Origin header, check it and skip other checks
	if origin := r.Header.Get("Origin"); origin != "" {
		u, _ := url.Parse(origin)
		return sameOrigin(u, p)
	}

	for _, header := range whitelistHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	referrer := r.Header.Get("Referer")

	u, _ := url.Parse(referrer)
	return sameOrigin(u, p)
}

// Check if two URLs have the same origin
func sameOrigin(u1, u2 *url.URL) bool {
	// host also contains the port if one was specified
	return (u1.Scheme == u2.Scheme && u1.Host == u2.Host)
}
