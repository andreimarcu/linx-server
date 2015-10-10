package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

type RespOkJSON struct {
	Filename   string
	Url        string
	Delete_Key string
	Expiry     string
	Size       string
}

type RespErrJSON struct {
	Error string
}

func TestSetup(t *testing.T) {
	Config.siteURL = "http://linx.example.org/"
	Config.filesDir = path.Join(os.TempDir(), generateBarename())
	Config.metaDir = Config.filesDir + "_meta"
	Config.maxSize = 1024 * 1024 * 1024
	Config.noLogs = true
	Config.siteName = "linx"
}

func TestIndex(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Click or Drop file") {
		t.Fatal("String 'Click or Drop file' not found in index response")
	}
}

func TestNotFound(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/url/should/not/exist", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestFileNotFound(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()

	req, err := http.NewRequest("GET", "/selif/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestDisplayNotFound(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()

	req, err := http.NewRequest("GET", "/"+filename, nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestPostCodeUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()
	extension := "txt"

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequest("POST", "/upload/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", Config.siteURL)

	mux.ServeHTTP(w, req)

	if w.Code != 303 {
		t.Fatalf("Status code is not 303, but %d", w.Code)
	}

	if w.Header().Get("Location") != "/"+filename+"."+extension {
		t.Fatalf("Was redirected to %s instead of /%s", w.Header().Get("Location"), filename)
	}
}

func TestPostCodeUploadWhitelistedHeader(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()
	extension := "txt"

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequest("POST", "/upload/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Linx-Expiry", "0")

	mux.ServeHTTP(w, req)

	if w.Code != 303 {
		t.Fatalf("Status code is not 303, but %d", w.Code)
	}
}

func TestPostCodeUploadNoReferrer(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()
	extension := "txt"

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequest("POST", "/upload/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mux.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Status code is not 400, but %d", w.Code)
	}
}

func TestPostCodeUploadBadOrigin(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()
	extension := "txt"

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", filename)
	form.Add("extension", extension)

	req, err := http.NewRequest("POST", "/upload/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", Config.siteURL)
	req.Header.Set("Origin", "http://example.com/")

	mux.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("Status code is not 400, but %d", w.Code)
	}
}

func TestPostCodeExpiryJSONUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	form := url.Values{}
	form.Add("content", "File content")
	form.Add("filename", "")
	form.Add("expires", "60")

	req, err := http.NewRequest("POST", "/upload/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.PostForm = form
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", Config.siteURL)

	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 200, but %d", w.Code)
	}

	var myjson RespOkJSON
	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	myExp, err := strconv.ParseInt(myjson.Expiry, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	curTime := time.Now().Unix()

	if myExp < curTime {
		t.Fatalf("File expiry (%d) is smaller than current time (%d)", myExp, curTime)
	}

	if myjson.Size != "12" {
		t.Fatalf("File size was not 12 but %s", myjson.Size)
	}
}

func TestPostUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}

	fw.Write([]byte("File content"))
	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 303 {
		t.Fatalf("Status code is not 303, but %d", w.Code)
	}

	if w.Header().Get("Location") != "/"+filename {
		t.Fatalf("Was redirected to %s instead of /%s", w.Header().Get("Location"), filename)
	}
}

func TestPostJSONUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}

	fw.Write([]byte("File content"))
	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 200, but %d", w.Code)
	}

	var myjson RespOkJSON
	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename != filename {
		t.Fatalf("Filename is not '%s' but '%s' ", filename, myjson.Filename)
	}

	if myjson.Expiry != "0" {
		t.Fatalf("File expiry is not 0 but %s", myjson.Expiry)
	}

	if myjson.Size != "12" {
		t.Fatalf("File size was not 12 but %s", myjson.Size)
	}
}

func TestPostExpiresJSONUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("File content"))

	exp, err := mw.CreateFormField("expires")
	if err != nil {
		t.Fatal(err)
	}
	exp.Write([]byte("60"))

	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 200, but %d", w.Code)
	}

	var myjson RespOkJSON
	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename != filename {
		t.Fatalf("Filename is not '%s' but '%s' ", filename, myjson.Filename)
	}

	myExp, err := strconv.ParseInt(myjson.Expiry, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	curTime := time.Now().Unix()

	if myExp < curTime {
		t.Fatalf("File expiry (%d) is smaller than current time (%d)", myExp, curTime)
	}

	if myjson.Size != "12" {
		t.Fatalf("File size was not 12 but %s", myjson.Size)
	}
}

func TestPostRandomizeJSONUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("File content"))

	rnd, err := mw.CreateFormField("randomize")
	if err != nil {
		t.Fatal(err)
	}
	rnd.Write([]byte("true"))

	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 200, but %d", w.Code)
	}

	var myjson RespOkJSON
	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename == filename {
		t.Fatalf("Filename (%s) is not random (%s)", filename, myjson.Filename)
	}

	if myjson.Size != "12" {
		t.Fatalf("File size was not 12 but %s", myjson.Size)
	}
}

func TestPostEmptyUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}

	fw.Write([]byte(""))
	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 500, but %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Empty file") {
		t.Fatal("Response did not contain 'Empty file'")
	}
}

func TestPostEmptyJSONUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".txt"

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}

	fw.Write([]byte(""))
	mw.Close()

	req, err := http.NewRequest("POST", "/upload/", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", Config.siteURL)
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Log(w.Body.String())
		t.Fatalf("Status code is not 500, but %d", w.Code)
	}

	var myjson RespErrJSON
	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Error != "Could not upload file: Empty file" {
		t.Fatal("Json 'error' was not 'Empty file' but " + myjson.Error)
	}
}

func TestPutUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	mux.ServeHTTP(w, req)

	if w.Body.String() != Config.siteURL+filename {
		t.Fatal("Response was not expected URL")
	}
}

func TestPutRandomizedUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Linx-Randomize", "yes")

	mux.ServeHTTP(w, req)

	if w.Body.String() == Config.siteURL+filename {
		t.Fatal("Filename was not random")
	}
}

func TestPutNoExtensionUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename()

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Linx-Randomize", "yes")

	mux.ServeHTTP(w, req)

	if w.Body.String() == Config.siteURL+filename {
		t.Fatal("Filename was not random")
	}
}

func TestPutEmptyUpload(t *testing.T) {
	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Linx-Randomize", "yes")

	mux.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Empty file") {
		t.Fatal("Response doesn't contain'Empty file'")
	}
}

func TestPutJSONUpload(t *testing.T) {
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	mux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename != filename {
		t.Fatal("Filename was not provided one but " + myjson.Filename)
	}
}

func TestPutRandomizedJSONUpload(t *testing.T) {
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Randomize", "yes")

	mux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	if myjson.Filename == filename {
		t.Fatal("Filename was not random ")
	}
}

func TestPutExpireJSONUpload(t *testing.T) {
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	filename := generateBarename() + ".ext"

	req, err := http.NewRequest("PUT", "/upload/"+filename, strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Expiry", "600")

	mux.ServeHTTP(w, req)

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
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/upload", strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	mux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequest("DELETE", "/"+myjson.Filename, nil)
	req.Header.Set("Linx-Delete-Key", myjson.Delete_Key)
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Status code was not 200, but " + strconv.Itoa(w.Code))
	}

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename, nil)
	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}

	// Make sure torrent is also gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename+"/torrent", nil)
	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}
}

func TestPutAndOverwrite(t *testing.T) {
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/upload", strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	mux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	// Overwrite it
	w = httptest.NewRecorder()
	req, err = http.NewRequest("PUT", "/upload/"+myjson.Filename, strings.NewReader("New file content"))
	req.Header.Set("Linx-Delete-Key", myjson.Delete_Key)
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Status code was not 200, but " + strconv.Itoa(w.Code))
	}

	// Make sure it's the new file
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/selif/"+myjson.Filename, nil)
	mux.ServeHTTP(w, req)

	if w.Code == 404 {
		t.Fatal("Status code was 404")
	}

	if w.Body.String() != "New file content" {
		t.Fatal("File did not contain 'New file content")
	}
}

func TestPutAndSpecificDelete(t *testing.T) {
	var myjson RespOkJSON

	mux := setup()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/upload", strings.NewReader("File content"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Linx-Delete-Key", "supersecret")

	mux.ServeHTTP(w, req)

	err = json.Unmarshal([]byte(w.Body.String()), &myjson)
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	w = httptest.NewRecorder()
	req, err = http.NewRequest("DELETE", "/"+myjson.Filename, nil)
	req.Header.Set("Linx-Delete-Key", "supersecret")
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Status code was not 200, but " + strconv.Itoa(w.Code))
	}

	// Make sure it's actually gone
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename, nil)
	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}

	// Make sure torrent is gone too
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/"+myjson.Filename+"/torrent", nil)
	mux.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Fatal("Status code was not 404, but " + strconv.Itoa(w.Code))
	}
}

func TestShutdown(t *testing.T) {
	os.RemoveAll(Config.filesDir)
	os.RemoveAll(Config.metaDir)
}
