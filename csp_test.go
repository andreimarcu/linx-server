package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zenazn/goji"
)

var testCSPHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'none'; style-src 'self';",
	"X-Frame-Options":         "SAMEORIGIN",
	"X-Content-Type-Options":  "nosniff",
}

func TestContentSecurityPolicy(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.Use(ContentSecurityPolicy(CSPOptions{
		policy: testCSPHeaders["Content-Security-Policy"],
		frame:  testCSPHeaders["X-Frame-Options"],
	}))

	goji.DefaultMux.ServeHTTP(w, req)

	for k, v := range testCSPHeaders {
		if w.HeaderMap[k][0] != v {
			t.Fatalf("%s header did not match expected value set by middleware", k)
		}
	}
}

// vim:set ts=8 sw=8 noet:
