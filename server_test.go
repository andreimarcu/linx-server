package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/zenazn/goji"
)

func TestSetup(t *testing.T) {
	Config.siteURL = "http://linx.example.org/"
	Config.filesDir = path.Join(os.TempDir(), randomString(10))
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

	if !strings.Contains(w.Body.String(), "Click or Drop file") {
		t.Fatal("String 'Click or Drop file' not found in index response")
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

func TestPutUpload(t *testing.T) {
	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	goji.DefaultMux.ServeHTTP(w, req)

	if w.Body.String() != Config.siteURL+filename {
		t.Fatal("Response was not expected URL")
	}
}

func TestPutRandomizedUpload(t *testing.T) {
	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Randomized-Barename", "yes")

	goji.DefaultMux.ServeHTTP(w, req)

	if w.Body.String() == Config.siteURL+filename {
		t.Fatal("Filename was not random")
	}
}

func TestPutEmptyUpload(t *testing.T) {
	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Randomized-Barename", "yes")

	goji.DefaultMux.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Empty file") {
		t.Fatal("Response doesn't contain'Empty file'")
	}
}

func TestPutJSONUpload(t *testing.T) {
	type RespJSON struct {
		Filename   string
		Url        string
		Delete_Key string
		Expiry     string
		Size       string
	}
	var myjson RespJSON

	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	goji.DefaultMux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename != filename {
		t.Fatal("Filename was not provided one but " + myjson.Filename)
	}
}

func TestPutRandomizedJSONUpload(t *testing.T) {
	type RespJSON struct {
		Filename   string
		Url        string
		Delete_Key string
		Expiry     string
		Size       string
	}
	var myjson RespJSON

	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Randomized-Barename", "yes")

	goji.DefaultMux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename == filename {
		t.Fatal("Filename was not random ")
	}
}

func TestPutExpireJSONUpload(t *testing.T) {
	type RespJSON struct {
		Filename   string
		Url        string
		Delete_Key string
		Expiry     string
		Size       string
	}
	var myjson RespJSON

	w := httptest.NewRecorder()

	filename := randomString(10) + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-File-Expiry", "600")

	goji.DefaultMux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	expiry, err := strconv.Atoi(myjson.Expiry)
	if err != nil {
		t.Fatal("Expiry was not an integer")
	}
	if expiry < 1 {
		t.Fatal("Expiry was not set")
	}
}

func TestPutAndDelete(t *testing.T) {
	type RespJSON struct {
		Filename   string
		Url        string
		Delete_Key string
		Expiry     string
		Size       string
	}
	var myjson RespJSON

	w := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/upload", strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	goji.DefaultMux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequest("DELETE", "/"+myjson.Filename, nil)
	req.Header.Set("X-Delete-Key", myjson.Delete_Key)
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Status code was not 200, but " + strconv.Itoa(w.Code))
	}

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename, nil)
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}

	// Make sure torrent is also gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename+"/torrent", nil)
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}
}

func TestPutAndSpecificDelete(t *testing.T) {
	type RespJSON struct {
		Filename   string
		Url        string
		Delete_Key string
		Expiry     string
		Size       string
	}
	var myjson RespJSON

	w := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/upload", strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Delete-Key", "supersecret")

	goji.DefaultMux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequest("DELETE", "/"+myjson.Filename, nil)
	req.Header.Set("X-Delete-Key", "supersecret")
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Status code was not 200, but " + strconv.Itoa(w.Code))
	}

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename, nil)
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}

	// Make sure torrent is gone too
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename+"/torrent", nil)
	goji.DefaultMux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}
}

func TestShutdown(t *testing.T) {
	os.RemoveAll(Config.filesDir)
	os.RemoveAll(Config.metaDir)
}
