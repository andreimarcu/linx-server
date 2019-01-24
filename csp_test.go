package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/zenazn/goji"
)

var testCSPHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'none'; style-src 'self';",
	"Referrer-Policy":         "strict-origin-when-cross-origin",
	"X-Frame-Options":         "SAMEORIGIN",
}

func TestContentSecurityPolicy(t *testing.T) {
	Config.siteURL = "http://linx.example.org/"
	Config.filesDir = path.Join(os.TempDir(), generateBarename())
	Config.metaDir = Config.filesDir + "_meta"
	Config.maxSize = 1024 * 1024 * 1024
	Config.noLogs = true
	Config.siteName = "linx"
	Config.selifPath = "selif"
	Config.contentSecurityPolicy = testCSPHeaders["Content-Security-Policy"]
	Config.referrerPolicy = testCSPHeaders["Referrer-Policy"]
	Config.xFrameOptions = testCSPHeaders["X-Frame-Options"]
	mux := setup()

	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.Use(ContentSecurityPolicy(CSPOptions{
		policy:         testCSPHeaders["Content-Security-Policy"],
		referrerPolicy: testCSPHeaders["Referrer-Policy"],
		frame:          testCSPHeaders["X-Frame-Options"],
	}))

	mux.ServeHTTP(w, req)

	for k, v := range testCSPHeaders {
		if w.HeaderMap[k][0] != v {
			t.Fatalf("%s header did not match expected value set by middleware", k)
		}
	}
}
