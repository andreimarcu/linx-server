package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/zenazn/goji"
)

var a = 0

func TestSetup(t *testing.T) {
	Config.siteURL = "http://linx.example.org/"
	Config.filesDir = "/tmp/" + randomString(10)
	Config.metaDir = Config.filesDir + "_meta"
	Config.noLogs = true
	Config.siteName = "linx"
	setup()
}

func TestIndex(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.DefaultMux.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "file-uploader") {
		t.Error("String 'file-uploader' not found in index response")
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/url/should/not/exist", nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestFileNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	filename := randomString(10)

	req, err := http.NewRequest("GET", "/selif/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestDisplayNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	filename := randomString(10)

	req, err := http.NewRequest("GET", "/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestShutdown(t *testing.T) {
	os.RemoveAll(Config.filesDir)
	os.RemoveAll(Config.metaDir)
}
