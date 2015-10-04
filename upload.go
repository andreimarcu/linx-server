package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.org/taruti/mimemagic"
	"github.com/dchest/uniuri"
	"github.com/zenazn/goji/web"
)

// Describes metadata directly from the user request
type UploadRequest struct {
	src            io.Reader
	filename       string
	expiry         int64 // Seconds until expiry, 0 = never
	randomBarename bool
	deletionKey    string // Empty string if not defined
}

// Metadata associated with a file as it would actually be stored
type Upload struct {
	Filename  string // Final filename on disk
	Size      int64
	Expiry    int64  // Unix timestamp of expiry, 0=never
	DeleteKey string // Deletion key, one generated if not provided
}

func uploadPostHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}
	uploadHeaderProcess(r, &upReq)

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, headers, err := r.FormFile("file")
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file.")
			return
		}
		defer file.Close()

		r.ParseForm()
		if r.Form.Get("randomize") == "true" {
			upReq.randomBarename = true
		}
		upReq.expiry = parseExpiry(r.Form.Get("expires"))
		upReq.src = file
		upReq.filename = headers.Filename
	} else {
		if r.FormValue("content") == "" {
			oopsHandler(c, w, r, RespHTML, "Could not upload file.")
			return
		}
		extension := r.FormValue("extension")
		if extension == "" {
			extension = "txt"
		}

		upReq.src = strings.NewReader(r.FormValue("content"))
		upReq.expiry = parseExpiry(r.FormValue("expires"))
		upReq.filename = r.FormValue("filename") + "." + extension
	}

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file: "+err.Error())
			return
		}

		http.Redirect(w, r, "/"+upload.Filename, 301)
	}

}

func uploadPutHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}
	uploadHeaderProcess(r, &upReq)

	defer r.Body.Close()
	upReq.filename = c.URLParams["name"]
	upReq.src = r.Body

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespPLAIN, "Could not upload file: "+err.Error())
			return
		}

		fmt.Fprintf(w, Config.siteURL+upload.Filename)
	}
}

func uploadRemote(c web.C, w http.ResponseWriter, r *http.Request) {
	if r.FormValue("url") == "" {
		http.Redirect(w, r, "/", 301)
		return
	}

	upReq := UploadRequest{}
	grabUrl, _ := url.Parse(r.FormValue("url"))

	resp, err := http.Get(grabUrl.String())
	if err != nil {
		oopsHandler(c, w, r, RespAUTO, "Could not retrieve URL")
		return
	}

	upReq.filename = filepath.Base(grabUrl.Path)
	upReq.src = resp.Body

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file: "+err.Error())
			return
		}

		http.Redirect(w, r, "/"+upload.Filename, 301)
	}
}

func uploadHeaderProcess(r *http.Request, upReq *UploadRequest) {
	// For legacy reasons
	if r.Header.Get("X-Randomized-Filename") == "yes" {
		upReq.randomBarename = true
	} else if r.Header.Get("X-Randomized-Barename") == "yes" {
		upReq.randomBarename = true
	}

	upReq.deletionKey = r.Header.Get("X-Delete-Key")

	// Get seconds until expiry. Non-integer responses never expire.
	expStr := r.Header.Get("X-File-Expiry")
	upReq.expiry = parseExpiry(expStr)

}

func processUpload(upReq UploadRequest) (upload Upload, err error) {
	// Determine the appropriate filename, then write to disk
	barename, extension := barePlusExt(upReq.filename)

	if upReq.randomBarename || len(barename) == 0 {
		barename = generateBarename()
	}

	var header []byte
	if len(extension) == 0 {
		// Pull the first 512 bytes off for use in MIME detection
		header = make([]byte, 512)
		n, err := upReq.src.Read(header)
		if n == 0 || err != nil {
			return upload, errors.New("Empty file")
		}
		header = header[:n]

		// Determine the type of file from header
		mimetype := mimemagic.Match("", header)

		// If the mime type is in our map, use that
		// otherwise just use "ext"
		if val, exists := mimeToExtension[mimetype]; exists {
			extension = val
		} else {
			extension = "ext"
		}
	}

	upload.Filename = strings.Join([]string{barename, extension}, ".")

	_, err = os.Stat(path.Join(Config.filesDir, upload.Filename))

	fileexists := err == nil
	for fileexists {
		counter, err := strconv.Atoi(string(barename[len(barename)-1]))
		if err != nil {
			barename = barename + "1"
		} else {
			barename = barename[:len(barename)-1] + strconv.Itoa(counter+1)
		}
		upload.Filename = strings.Join([]string{barename, extension}, ".")

		_, err = os.Stat(path.Join(Config.filesDir, upload.Filename))
		fileexists = err == nil
	}

	dst, err := os.Create(path.Join(Config.filesDir, upload.Filename))
	if err != nil {
		return
	}
	defer dst.Close()

	// Get the rest of the metadata needed for storage
	upload.Expiry = getFutureTimestamp(upReq.expiry)

	// If no delete key specified, pick a random one.
	if upReq.deletionKey == "" {
		upload.DeleteKey = uniuri.NewLen(30)
	} else {
		upload.DeleteKey = upReq.deletionKey
	}

	metadataWrite(upload.Filename, &upload)

	bytes, err := io.Copy(dst, io.MultiReader(bytes.NewReader(header), upReq.src))
	if bytes == 0 {
		os.Remove(path.Join(Config.filesDir, upload.Filename))
		os.Remove(path.Join(Config.metaDir, upload.Filename))
		return upload, errors.New("Empty file")

	} else if err != nil {
		os.Remove(path.Join(Config.filesDir, upload.Filename))
		os.Remove(path.Join(Config.metaDir, upload.Filename))
		return
	}

	upload.Size = bytes
	return
}

func generateBarename() string {
	return uniuri.NewLenChars(8, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

func generateJSONresponse(upload Upload) []byte {
	js, _ := json.Marshal(map[string]string{
		"url":        Config.siteURL + upload.Filename,
		"filename":   upload.Filename,
		"delete_key": upload.DeleteKey,
		"expiry":     strconv.FormatInt(int64(upload.Expiry), 10),
		"size":       strconv.FormatInt(upload.Size, 10),
	})

	return js
}

var barePlusRe = regexp.MustCompile(`[^A-Za-z0-9\-]`)

func barePlusExt(filename string) (barename, extension string) {

	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)

	extension = path.Ext(filename)
	barename = filename[:len(filename)-len(extension)]

	extension = barePlusRe.ReplaceAllString(extension, "")
	barename = barePlusRe.ReplaceAllString(barename, "")

	return
}

func parseExpiry(expStr string) int64 {
	if expStr == "" {
		return 0
	} else {
		expiry, err := strconv.ParseInt(expStr, 10, 64)
		if err != nil {
			return 0
		} else {
			return int64(expiry)
		}
	}
}
