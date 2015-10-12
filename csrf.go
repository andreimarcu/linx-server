package main

import (
	"net/http"
	"strings"
)

func strictReferrerCheck(r *http.Request, prefix string, whitelistHeaders []string) bool {
	p := strings.TrimSuffix(prefix, "/")
	if origin := r.Header.Get("Origin"); origin != "" {
		// if there's an Origin header, check it and ignore the rest
		return strings.HasPrefix(origin, p)
	}

	for _, header := range whitelistHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	if referrer := r.Header.Get("Referer"); !strings.HasPrefix(referrer, p) {
		return false
	}

	return true
}
