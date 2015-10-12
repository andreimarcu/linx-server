package main

import (
	"net/http"
	"strings"
)

func strictReferrerCheck(r *http.Request, prefix string, whitelistHeaders []string) bool {
	for _, header := range whitelistHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}

	p := strings.TrimSuffix(prefix, "/")

	if referrer := r.Header.Get("Referer"); !strings.HasPrefix(referrer, p) {
		return false
	}

	if origin := r.Header.Get("Origin"); origin != "" && !strings.HasPrefix(origin, p) {
		return false
	}

	return true
}
