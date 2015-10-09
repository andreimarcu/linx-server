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

	if referrer := r.Header.Get("Referer"); !strings.HasPrefix(referrer, prefix) {
		return false
	}

	if origin := r.Header.Get("Origin"); origin != "" && !strings.HasPrefix(origin, prefix) {
		return false
	}

	return true
}
